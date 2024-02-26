package main

import (
	"fmt"
	"strconv"
	"strings"
)

const TERMINATOR = "\r\n"

type RespT byte

const (
	String RespT = '+'
	Bulk   RespT = '$'
	Array  RespT = '*'
	Null   RespT = '_'
)

type Resp struct {
	Type   RespT
	Value  []string
	Length int
	raw    []byte
}

func NewResp(data []byte, t ...RespT) *Resp {
	r := new(Resp)
	r.parse(data)
	if len(t) > 0 {
		r.Type = RespT(t[0])
	}
	return r
}

func (r *Resp) parse(raw []byte) {
	splitted := strings.Split(string(raw), TERMINATOR)
	if len(raw) > 0 {
		r.Type = RespT(splitted[0][0])
		fmt.Println(r.Type, " parse")
		r.Value = splitted[1:]
		r.raw = raw
	}

	if r.Type == Array || r.Type == Bulk {
		count, err := strconv.Atoi(splitted[0][1:])
		if err == nil {
			r.Length = count
		} else {
			r.Length = len(r.Value)
		}
	}
}

func (r *Resp) SetValue(values ...string) {
	if r.Value != nil {
		r.Value = append(r.Value, values...)
	} else {
		r.Value = values
	}

	if r.Type == Bulk {
		r.Length = len(r.Value[0])
	}
}

func (r *Resp) SetPong() {
	r.Type = String
	r.SetValue("PONG")
}

func (r *Resp) SetOK() {
	r.Type = String
	r.SetValue("OK")
}

// appends bulk string to array, parameters must be of string, int or bool type
func (r *Resp) AppendBulk(strs ...any) error {
	if r.Type != Array {
		return fmt.Errorf("response type should be array")
	}

	for _, v := range strs {
		var value string

		// TODO add other types
		switch temp := v.(type) {
		case int:
			value = strconv.Itoa(temp)
		case bool:
			value = strconv.FormatBool(temp)
		case string:
			value = temp
		default:
			return fmt.Errorf("invalid type in passed arguments")
		}

		r.Value = append(r.Value, fmt.Sprintf("$%v", len(value)), value)
		r.Length = len(r.Value) / 2
	}

	return nil
}

func (r *Resp) Parse() error {
	fmt.Printf("r.Type: %v\n", r.Type)
	if r.Type == 0 {
		return fmt.Errorf("resp must have a type")
	}

	if r.Length == 0 {
		r.Length = len(r.Value) / 2
	}

	r.raw = append(r.raw, byte(r.Type))
	if r.Type == Array || (r.Type == Bulk && !strings.HasPrefix(r.Value[0], "-")) {
		r.raw = append(r.raw, []byte(strconv.Itoa(r.Length))...)
		r.raw = append(r.raw, []byte(TERMINATOR)...)
	}

	for _, v := range r.Value {
		r.raw = append(r.raw, []byte(v)...)
		r.raw = append(r.raw, []byte(TERMINATOR)...)
	}

	return nil
}

func (r Resp) Bytes() []byte {
	return r.raw
}
