package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"regexp"
	"simpleKV/resp"
	"strconv"
	"strings"
)

type IClient interface {
	Close() error
	Set(k string, v any) error
	Get(k string) (any, error)
	Del(k string) error
	Command(arg string) error
	Info() error
	Scan(cursor int, matchPattern *regexp.Regexp, count int) ([]string, int, error)
}

type client struct {
	conn net.Conn
}

func NewClient(address string) (IClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("could not connect to server: %v", err)
	}

	return &client{conn: conn}, nil
}

func (c *client) Close() error {
	return c.conn.Close()
}

func (c *client) Set(k string, v any) error {
	value := fmt.Sprintf("%v", v)
	command := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
		len(k), k, len(value), value)
	_, err := c.conn.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("could not send SET command: %v", err)
	}

	response, err := c.readResponse()
	if err != nil {
		return err
	}

	if response.Type == resp.SIMPLE_STRING && response.String == "OK" {
		return nil
	}

	return errors.New("SET command failed")
}

func (c *client) Get(k string) (any, error) {
	command := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(k), k)
	_, err := c.conn.Write([]byte(command))
	if err != nil {
		return nil, fmt.Errorf("could not send GET command: %v", err)
	}

	response, err := c.readResponse()
	if err != nil {
		return nil, err
	}

	if response.Type == resp.BULK_STRING {
		return response.BulkString, nil
	} else if response.Type == resp.NULL {
		return nil, nil
	}

	return nil, errors.New("GET command failed or returned unexpected type")
}

func (c *client) Del(k string) error {
	command := fmt.Sprintf("*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n", len(k), k)
	_, err := c.conn.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("could not send DEL command: %v", err)
	}

	response, err := c.readResponse()
	if err != nil {
		return err
	}

	if response.Type == resp.INTEGER && response.Integer > 0 {
		return nil
	}

	return errors.New("DEL command failed or key not found")
}

func (c *client) Command(arg string) error {
	command := fmt.Sprintf("*2\r\n$7\r\nCOMMAND\r\n$%d\r\n%s\r\n", len(arg), arg)
	_, err := c.conn.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("could not send COMMAND command: %v", err)
	}

	response, err := c.readResponse()
	if err != nil {
		return err
	}

	if response.Type == resp.ARRAY {
		fmt.Println("COMMAND response:", response.Array)
		return nil
	}

	return errors.New("COMMAND failed or returned unexpected type")
}

func (c *client) Info() error {
	command := "*1\r\n$4\r\nINFO\r\n"
	_, err := c.conn.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("could not send INFO command: %v", err)
	}

	response, err := c.readResponse()
	if err != nil {
		return err
	}

	if response.Type == resp.SIMPLE_STRING {
		fmt.Println("INFO response:\n", response.String)
		return nil
	}

	return errors.New("INFO command failed or returned unexpected type")
}

func (c *client) Scan(cursor int, matchPattern *regexp.Regexp, count int) ([]string, int, error) {
	command := fmt.Sprintf("*4\r\n$4\r\nSCAN\r\n$%d\r\n%d\r\n$5\r\nMATCH\r\n$%d\r\n%s\r\n$5\r\nCOUNT\r\n$%d\r\n%d\r\n",
		len(strconv.Itoa(cursor)), cursor,
		len(matchPattern.String()), matchPattern.String(),
		len(strconv.Itoa(count)), count)

	_, err := c.conn.Write([]byte(command))
	if err != nil {
		return nil, 0, fmt.Errorf("could not send SCAN command: %v", err)
	}

	response, err := c.readResponse()
	if err != nil {
		return nil, 0, err
	}

	if response.Type == resp.ARRAY && len(response.Array) == 2 {
		nextCursor, err := strconv.Atoi(response.Array[0].BulkString)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor in SCAN response: %v", err)
		}

		keys := []string{}
		if response.Array[1].Type == resp.ARRAY {
			for _, v := range response.Array[1].Array {
				if v.Type == resp.BULK_STRING {
					keys = append(keys, v.BulkString)
				}
			}
		}

		return keys, nextCursor, nil
	}

	return nil, 0, errors.New("SCAN command failed or returned unexpected type")
}

func (c *client) readResponse() (resp.Value, error) {
	reader := bufio.NewReader(c.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return resp.Value{}, fmt.Errorf("could not read response: %v", err)
	}

	line = strings.TrimSpace(line)

	switch line[0] {
	case '+':
		return resp.Value{Type: resp.SIMPLE_STRING, String: line[1:]}, nil
	case '-':
		return resp.NewErrorValue(line[1:]), nil
	case ':':
		val, err := strconv.ParseInt(line[1:], 10, 64)
		if err != nil {
			return resp.Value{}, fmt.Errorf("invalid integer response: %v", err)
		}
		return resp.Value{Type: resp.INTEGER, Integer: val}, nil
	case '$':
		length, err := strconv.Atoi(line[1:])
		if err != nil || length < 0 {
			return resp.Value{Type: resp.NULL}, nil
		}
		data := make([]byte, length+2)
		_, err = reader.Read(data)
		if err != nil {
			return resp.Value{}, fmt.Errorf("could not read bulk string: %v", err)
		}
		return resp.Value{Type: resp.BULK_STRING, BulkString: string(data[:length])}, nil
	case '*':
		count, err := strconv.Atoi(line[1:])
		if err != nil || count < 0 {
			return resp.Value{}, fmt.Errorf("invalid array response: %v", err)
		}
		array := []resp.Value{}
		for i := 0; i < count; i++ {
			elem, err := c.readResponse()
			if err != nil {
				return resp.Value{}, err
			}
			array = append(array, elem)
		}
		return resp.Value{Type: resp.ARRAY, Array: array}, nil
	default:
		return resp.Value{}, fmt.Errorf("unknown response type: %v", line)
	}
}
