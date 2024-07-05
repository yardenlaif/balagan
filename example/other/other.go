package other

type A interface {
	Floo()
}

var a interface{}

func B() {
	if b, b := a.(A); b {
		c(&b)
	}
}

func c(d *A) {
}
