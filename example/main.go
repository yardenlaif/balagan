package main

import (
	"fmt"
	"io"
)

type myInterface interface {
	Hey()
}

type t struct{}

func (t *t) Close() error {
	return nil
}

func main() {
	var1 := 1
	var2 := "hello"
	var v io.Closer = &t{}
	v.Close()
	fmt.Printf("%d, %s\n", var1, var2)
}
