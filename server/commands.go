package server

import (
	"fmt"
	"simpleKV/resp"
	"strconv"
	"strings"
)

func (s *server) handleRequest(req resp.Value) resp.Value {
	if req.Type != resp.ARRAY || len(req.Array) < 1 {
		return resp.NewErrorValue("ERR invalid request format")
	}

	cmd := strings.ToUpper(req.Array[0].BulkString)

	switch cmd {
	case "SET":
		if len(req.Array) != 3 {
			return resp.NewErrorValue("ERR wrong number of arguments for 'SET' command")
		}
		key := req.Array[1].BulkString
		value := req.Array[2]
		s.store.Set(key, value)
		return resp.Value{Type: resp.SIMPLE_STRING, String: "OK"}

	case "GET":
		if len(req.Array) != 2 {
			return resp.NewErrorValue("ERR wrong number of arguments for 'GET' command")
		}
		key := req.Array[1].BulkString
		value, found := s.store.Get(key)
		if !found {
			return resp.Value{Type: resp.NULL}
		}
		return value

	case "DEL":
		if len(req.Array) < 2 {
			return resp.NewErrorValue("ERR wrong number of arguments for 'DEL' command")
		}
		deletedCount := 0
		for _, keyValue := range req.Array[1:] {
			if s.store.Del(keyValue.BulkString) {
				deletedCount++
			}
		}
		return resp.NewIntegerValue(int64(deletedCount))

	case "COMMAND":
		if len(req.Array) == 2 && strings.ToUpper(req.Array[1].BulkString) == "DOCS" {
			return resp.Value{
				Type:  resp.ARRAY,
				Array: []resp.Value{},
			}
		}
		return resp.NewErrorValue("ERR unknown subcommand for 'COMMAND'")

	case "INFO":
		info := "redis_version: 0.0.1\n"
		info += "connected_clients: 1\n"
		info += "used_memory: 1024\n"
		info += "uptime_in_seconds: 3600\n"
		info += "keys_count: 1000\n"

		return resp.Value{
			Type:   resp.SIMPLE_STRING,
			String: info,
		}

	case "SCAN":
		cursor := "0"
		matchPattern := ""
		count := 100

		if len(req.Array) > 1 {
			cursor = req.Array[1].BulkString
		}
		if len(req.Array) > 2 {
			matchPattern = req.Array[2].BulkString
		}
		if len(req.Array) > 3 {
			parsedCount, err := strconv.Atoi(req.Array[3].BulkString)
			if err == nil {
				count = parsedCount
			}
		}

		result := s.store.Scan(cursor, matchPattern, count)

		return result

	default:
		return resp.NewErrorValue(fmt.Sprintf("ERR unknown command '%s'", cmd))
	}
}
