package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type OP_CODE byte

const (
	AUX            OP_CODE = 0xFA
	RESIZE_DB      OP_CODE = 0xFB
	EXPIRE_TIME_MS OP_CODE = 0xFC
	EXPIRE_TIME    OP_CODE = 0xFD
	SELECT_DB      OP_CODE = 0xFE
	EOF            OP_CODE = 0xFF
)

// rdb dosyasını nasıl okuyabileceğimi anlamadım, phind e falan da sordum ama gene anlamadım
// sayısal tasarımı bitirdikten sonra tekrar dönecem, o zaman da codecrafts ücrestsiz sürümü bitmiş olacak
// ama dosya okumayı ve değerleri kaydetmeyi yapsam yeter gibi
// o zamana kadar
// TODO implement this function
func ReadRdbFile(db *map[string]string) error {
	f, err := os.Open("dump.rdb")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
		}
		fmt.Printf("---file---\n%v\n", line)
	}

	return nil
}
