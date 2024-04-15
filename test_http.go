package main

import (
	"bytes"
	"fmt"
)

func main() {
	b := bytes.NewBuffer(nil)
	b.Write([]byte("123123"))
	fmt.Println(b.String())

	//NOTE:
}
