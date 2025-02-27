package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type IReader interface {
	Read() (Value, error)
}

type reader struct {
	rd *bufio.Reader
}

func NewReader(rd io.Reader) IReader {
	return &reader{
		rd: bufio.NewReader(rd),
	}
}

func (r *reader) Read() (Value, error) {
	_type, err := r.rd.ReadByte()

	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK_STRING:
		return r.readBulkString()
	case INTEGER:
		return r.readInteger()
	case SIMPLE_STRING:
		return r.readSimpleString()
	case SIMPLE_ERROR:
		return r.readSimpleError()
	case NULL:
		return r.readNull()
	case BOOLEAN:
		return r.readBoolean()
	case DOUBLE:
		return r.readDouble()
	case BIG_NUMBER:
		return r.readBigNumber()
	case VERBATIM_STRING:
		return r.readVerbatimString()
	case MAP:
		return r.readMap()
	case ATTRIBUTE:
		return r.readAttribute()
	case SET:
		return r.readSet()
	case PUSH:
		return r.readPush()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *reader) readArray() (Value, error) {
	v := Value{Type: ARRAY}

	length, err := r.readLength()
	if err != nil {
		return v, err
	}

	if length < 0 {
		return v, fmt.Errorf("Array length cant be negative")
	}

	v.Array = make([]Value, length)
	for i := range length {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		v.Array[i] = val
	}

	return v, nil
}

func (r *reader) readBulkString() (Value, error) {
	v := Value{Type: BULK_STRING}

	length, err := r.readLength()
	if err != nil {
		return v, err
	}

	if length < 0 {
		return v, fmt.Errorf("Bulk length cant be negative")
	}

	bulk := make([]byte, length)

	totalRead := 0
	for totalRead < length {
		n, err := r.rd.Read(bulk[totalRead:])
		if err != nil {
			return v, fmt.Errorf("Error reading bulk data: %v", err)
		}
		if n == 0 {
			return v, fmt.Errorf("Unexpected EOF: read %d bytes, expected %d bytes", totalRead, length)
		}
		totalRead += n
	}

	v.BulkString = string(bulk)

	line, _, err := r.readLine()
	if err != nil {
		return v, fmt.Errorf("Error reading trailing CRLF: %v", err)
	}
	if string(line) != "" {
		return v, fmt.Errorf("Expected CRLF after bulk data, but got '%s'", line)
	}

	return v, nil
}

func (r *reader) readInteger() (Value, error) {
	v := Value{Type: INTEGER}

	line, _, err := r.readLine()

	n, err := strconv.Atoi(strings.TrimSpace(string(line)))
	if err != nil {
		return v, err
	}

	v.Integer = int64(n)
	return v, nil
}

func (r *reader) readSimpleString() (Value, error) {
	line, _, err := r.readLine()

	v := Value{
		Type:   SIMPLE_STRING,
		String: string(line),
	}

	if err != nil {
		return v, err
	}

	return v, nil
}

func (r *reader) readSimpleError() (Value, error) {
	v := Value{Type: SIMPLE_ERROR}

	line, _, err := r.readLine()

	if err != nil {
		return v, err
	}

	v.String = string(line)
	return v, nil
}

func (r *reader) readNull() (Value, error) {
	v := Value{Type: NULL}

	_, _, err := r.readLine()
	if err != nil {
		return v, err
	}

	return v, nil
}

func (r *reader) readBoolean() (Value, error) {
	v := Value{
		Type:    BOOLEAN,
		Boolean: false,
	}

	line, _, err := r.readLine()
	if err != nil {
		return v, err
	}

	if len(line) != 1 || (line[0] != 'f' && line[0] != 't') {
		return v, fmt.Errorf("Invalid Boolean")
	}

	b := line[0] == 't'
	v.Boolean = b

	return v, nil
}

func (r *reader) readDouble() (Value, error) {
	v := Value{Type: DOUBLE}

	line, _, err := r.readLine()

	n, err := strconv.ParseFloat(strings.TrimSpace(string(line)), 64)
	if err != nil {
		return v, err
	}

	v.Double = n
	return v, nil
}

func (r *reader) readBigNumber() (Value, error) {
	v := Value{Type: BIG_NUMBER}

	line, _, err := r.readLine()
	if err != nil {
		return v, err
	}

	v.String = string(line)
	return v, nil
}

func (r *reader) readVerbatimString() (Value, error) {
	v := Value{Type: VERBATIM_STRING}

	length, err := r.readLength()
	if err != nil {
		return v, err
	}

	if length < 4 {
		return v, fmt.Errorf("Verbatim string length too short")
	}

	data := make([]byte, length)
	_, err = io.ReadFull(r.rd, data)
	if err != nil {
		return v, err
	}

	if !strings.HasPrefix(string(data), "txt:") {
		return v, fmt.Errorf("Invalid verbatim string format, expected 'txt:' prefix")
	}

	v.String = string(data[4:])
	_, _, err = r.readLine()
	if err != nil {
		return v, err
	}

	return v, nil
}

func (r *reader) readMap() (Value, error) {
	v := Value{Type: MAP}

	length, err := r.readLength()
	if err != nil {
		return v, err
	}

	if length < 0 {
		return v, fmt.Errorf("Map length can't be negative")
	}

	v.Array = make([]Value, length*2) // Map consists of key-value pairs
	for i := range length*2 {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		v.Array[i] = val
	}

	return v, nil
}

func (r *reader) readAttribute() (Value, error) {
	v := Value{Type: ATTRIBUTE}

	length, err := r.readLength()
	if err != nil {
		return v, err
	}

	if length < 0 {
		return v, fmt.Errorf("Attribute length can't be negative")
	}

	v.Array = make([]Value, length*2) // Key-value pairs
	for i := range length * 2 {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		v.Array[i] = val
	}

	return v, nil
}

func (r *reader) readSet() (Value, error) {
	v := Value{Type: SET}

	length, err := r.readLength()
	if err != nil {
		return v, err
	}

	if length < 0 {
		return v, fmt.Errorf("Set length can't be negative")
	}

	v.Array = make([]Value, length)
	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		v.Array[i] = val
	}

	return v, nil
}

func (r *reader) readPush() (Value, error) {
	v := Value{Type: PUSH}

	length, err := r.readLength()
	if err != nil {
		return v, err
	}

	if length < 0 {
		return v, fmt.Errorf("Push length can't be negative")
	}

	v.Array = make([]Value, length)
	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		v.Array[i] = val
	}

	return v, nil
}

func (r *reader) readLength() (int, error) {
	line, _, err := r.readLine()
	if err != nil {
		return 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i64), nil
}

func (r *reader) readLine() (line []byte, numBytes int, err error) {
	for {
		b, err := r.rd.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		numBytes += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}

	return line[:len(line)-2], numBytes, nil
}
