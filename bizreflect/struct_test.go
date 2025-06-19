package bizreflect

import (
	"testing"
)

type example struct {
	Name   string `biz:"123"`
	Age    int64  `biz:"10"`
	Status struct {
		Active string `biz:"active"`
	}
	Info struct {
		Name string
		Age  int64 `biz:"11"`
	} `Name:"name"`
}

func Test123(t *testing.T) {
	a := example{}
	t.Log(InitStruct(&a))
	t.Log(a)
}
