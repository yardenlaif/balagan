package example

import (
	"fmt"
	"io"
)

type i interface {
	foo()
}

type iImpl struct {
	f int
}

func (iyarden *iImpl) foo() {
}
func (iyarden *iImpl) Floo() {
}

func (iyarden *iImpl) Read(p []byte) (n int, err error) {
	return 0, nil
}

func a(iyarden *iImpl, reader io.Reader) {
	var b string
	c := 1
	b = "%d"
	fmt.Printf(b, c)
}

func d() {
	a(nil, &iImpl{})
}
