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
	var t other.A
	_ = t
	return 0, nil
}

func j(k *b, l io.Reader) {
	var m string
	n := 1
	m = "%d"
	fmt.Printf(m, n)
}

func o() {
	j(nil, &b{})
}
