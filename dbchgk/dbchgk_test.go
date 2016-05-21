package dbchgk

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestDbKAND(t *testing.T) {
	data, _ := ioutil.ReadFile("../testdata/kand.html")
	s, _ := dbKAND(data)
	for _, q := range s.Questions {
		fmt.Println(q.Picture)
	}
}

func TestDbCustom(t *testing.T) {
	data, _ := ioutil.ReadFile("../testdata/custom.html")
	s, _ := dbCustom("http://kane1.ipq.co/chgk-files/Uhodyaschaya_natura-2016/Uhodyaschaya_natura-2016.html", data)
	for _, q := range s.Questions {
		fmt.Println(">>>Question", q.Question)
		fmt.Println(">>>Picture", q.Picture)
		fmt.Println(">>>Answer", q.Answer)
	}
}
