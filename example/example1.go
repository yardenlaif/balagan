package example

import (
	"fmt"
	"io"

	"github.com/YardenLaif/balagan/example/other"
)

type a interface {
	foo()
}

type b struct {
	c int
}

func (d *b) foo() {
}
func (e *b) Floo() {
}

func (f *b) Read(g []byte) (h int, i error) {
	return 0, nil
}

func j(k *b, l io.Reader) {
	var m string
	n := 1
	m = "%d"
	k.c = 3
	var o interface{}
	switch something := o.(type) {
	case int:
		fmt.Printf("is int: %v\n", something)
	case float32:
		fmt.Printf("is int: %v\n", something)
	case io.ReadCloser:
		something.Close()
	}
	fmt.Printf(m, n)
	other.B()
}

func p() {
	j(nil, &b{})
}
