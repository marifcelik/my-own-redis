package main

type RespT byte

const (
	Integer RespT = ':'
	String  RespT = '+'
	Bulk    RespT = '$'
	Array   RespT = '*'
)

type Resp struct {
	Type  RespT
	Value string
	Raw   []byte
	Count int
}

func NewResp(data []byte) *Resp {
	r := new(Resp)
	r.Raw = data
	r.parse()
	return r
}

// TODO create some methods for reading response and writing response
func (r *Resp) parse() {

}
