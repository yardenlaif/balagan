package other

type A interface {
	Floo()
}

var a interface{}

type b int8

func (c *b) d() {
}

func (f *b) Read(g []byte) (h int, i error) {
	return 0, nil
}

func B() {
	if _, e := a.(A); e {
		f()
	}
}

func f() {
}
