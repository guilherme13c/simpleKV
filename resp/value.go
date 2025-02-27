package resp

import (
	"fmt"
	"strconv"
)

type Value struct {
	Type       RESPType
	String     string
	Integer    int64
	Double     float64
	BulkString string
	Boolean    bool
	Array      []Value
}

func (v Value) Marshal() []byte {
	switch v.Type {
	case ARRAY:
		return v.marshalArray()
	case BULK_STRING:
		return v.marshalBulk()
	case INTEGER:
		return v.marshalNum()
	case SIMPLE_STRING:
		return v.marshalString()
	case SIMPLE_ERROR:
		return v.marshallError()
	case NULL:
		return v.marshallNull()
	case BOOLEAN:
		return v.marshalBoolean()
	case DOUBLE:
		return v.marshalDouble()
	case BIG_NUMBER:
		return v.marshalBigNumber()
	case VERBATIM_STRING:
		return v.marshalVerbatimString()
	case MAP:
		return v.marshalMap()
	case ATTRIBUTE:
		return v.marshalAttribute()
	case SET:
		return v.marshalSet()
	case PUSH:
		return v.marshalPush()
	default:
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, SIMPLE_STRING)
	bytes = append(bytes, v.String...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalNum() []byte {

	var bytes []byte

	bytes = append(bytes, INTEGER)
	bytes = append(bytes, strconv.FormatInt(int64(v.Integer), 10)...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK_STRING)
	bytes = append(bytes, strconv.Itoa(len(v.BulkString))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, []byte(v.BulkString)...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	len := len(v.Array)
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len; i++ {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, SIMPLE_ERROR)
	bytes = append(bytes, []byte(v.String)...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}

func (v Value) marshalBoolean() []byte {
	var bytes []byte
	bytes = append(bytes, BOOLEAN)
	if v.Boolean {
		bytes = append(bytes, 't')
	} else {
		bytes = append(bytes, 'f')
	}
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalDouble() []byte {
	var bytes []byte
	bytes = append(bytes, DOUBLE)
	bytes = append(bytes, strconv.FormatFloat(v.Double, 'f', -1, 64)...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalBigNumber() []byte {
	var bytes []byte
	bytes = append(bytes, BIG_NUMBER)
	bytes = append(bytes, v.String...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalVerbatimString() []byte {
	var bytes []byte
	bytes = append(bytes, VERBATIM_STRING)
	bytes = append(bytes, strconv.Itoa(len(v.String)+4)...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, "txt:"...)
	bytes = append(bytes, v.String...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalMap() []byte {
	var bytes []byte
	len := len(v.Array) / 2
	bytes = append(bytes, MAP)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := range len*2 {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshalAttribute() []byte {
	var bytes []byte
	len := len(v.Array) / 2
	bytes = append(bytes, ATTRIBUTE)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := range len*2 {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshalSet() []byte {
	var bytes []byte
	len := len(v.Array)
	bytes = append(bytes, SET)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := range len {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshalPush() []byte {
	var bytes []byte
	len := len(v.Array)
	bytes = append(bytes, PUSH)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := range len {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}

	return bytes
}

func NewSetValue(key, value string) Value {
	arr := []Value{{Type: BULK_STRING, BulkString: "set"}, {Type: BULK_STRING, BulkString: key}, {Type: BULK_STRING, BulkString: value}}
	val := Value{Type: ARRAY, Array: arr}

	return val
}

func NewHsetValue(hash, key, value string) Value {
	fmt.Println(hash)
	arr := []Value{{Type: BULK_STRING, BulkString: "hset"}, {Type: BULK_STRING, BulkString: hash}, {Type: BULK_STRING, BulkString: key}, {Type: BULK_STRING, BulkString: value}}
	val := Value{Type: ARRAY, Array: arr}

	return val
}

func NewDelValue(keys []string) Value {
	arr := []Value{{Type: BULK_STRING, BulkString: "del"}}

	for _, key := range keys {
		v := Value{Type: BULK_STRING, BulkString: key}

		arr = append(arr, v)
	}
	val := Value{Type: ARRAY, Array: arr}

	return val
}

func NewErrorValue(message string) Value {

	val := Value{Type: SIMPLE_ERROR, String: message}

	return val
}

func NewIntegerValue(number int64) Value {

	val := Value{Type: INTEGER, Integer: number}

	return val
}
