package other

type Interface interface {
	Floo()
}

var f interface{}

func A() {
	if _, ok := f.(Interface); ok {
		a()
	}
}

func a() {
}
