package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/denisbrodbeck/machineid"
	"log"
	"strconv"
)

func main() {
	num := int64(999)
	hex := []byte(strconv.FormatInt(num, 16))
	data := make([]byte, 32-len(hex), 32)
	data = append(data, hex...)
	spew.Dump(data)
	id, err := machineid.ID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
}
