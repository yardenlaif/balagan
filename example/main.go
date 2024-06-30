package main

import (
	"fmt"
	"io"
)

type myInterface interface {
	Hey()
}

type veryUnique struct{}

func (a *veryUnique) Close() error {
	return nil
}

func main() {
	if newGlobal == "sd" {
		// sdflkj
		callMe()
	}
	var1 := "hey"
	var impls io.Closer
	impls = &veryUnique{}
	fmt.Printf("%v %v\n", var1, impls)
}
