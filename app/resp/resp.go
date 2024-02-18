package resp

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
	Raw    []byte
	Length int
}

func NewResp(data []byte, t ...RespT) *Resp {
	fmt.Println(string(data))
	r := new(Resp)
	r.Raw = data
	r.parse()
	if len(t) > 0 {
		r.Type = RespT(t[0])
	}
	return r
}

// i have not figured out how can i parse the resp string
func (r *Resp) parse() {
	splitted := strings.Split(string(r.Raw), TERMINATOR)
	if len(r.Raw) > 0 {
		r.Type = RespT(splitted[0][0])
		r.Value = splitted[1:]
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
		r.Length = len(r.Value)
	}

	return nil
}

func (r *Resp) Parse() error {
	var raw []byte

	if r.Type == 0 {
		return fmt.Errorf("resp must have a type")
	}

	raw = append(raw, byte(r.Type))
	if r.Type == Array || (r.Type == Bulk && !strings.HasPrefix(r.Value[0], "-")) {
		raw = append(raw, []byte(strconv.Itoa(r.Length))...)
		raw = append(raw, []byte(TERMINATOR)...)
	}

	for _, v := range r.Value {
		raw = append(raw, []byte(v)...)
		raw = append(raw, []byte(TERMINATOR)...)
	}

	r.Raw = raw
	return nil
}

func (r Resp) Bytes() []byte {
	return r.Raw
}
