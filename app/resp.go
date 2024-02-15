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

// TODO create some methods for reading response and writing response
