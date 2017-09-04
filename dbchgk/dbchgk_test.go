package dbchgk

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestDbCHGK(t *testing.T) {
	data, _ := ioutil.ReadFile("../testdata/chgk.info.xml")
	s, _ := dbCHGK(data)
	for _, q := range s.Questions {
		fmt.Println(q.Picture)
	}
}

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

func TestDbCustom1(t *testing.T) {
	data, _ := ioutil.ReadFile("../testdata/custom1.html")
	s, _ := dbCustom("http://kane1.ipq.co/chgk-files/Uhodyaschaya_natura-2016/Uhodyaschaya_natura-2016.html", data)
	for _, q := range s.Questions {
		fmt.Println(">>>Question", q.Question)
		fmt.Println(">>>Picture", q.Picture)
		fmt.Println(">>>Answer", q.Answer)
	}
}

func TestDbRegexp(t *testing.T) {
	data, _ := ioutil.ReadFile("../testdata/regexp.html")
	//!startregex http://faig-huseynov.livejournal.com/618.html 1 ~~j-c-resize-images..\s*\d+(.*?)Ответ:~~~~Ответ:(.*?)Автор:
	//!startregex http://kane1.ipq.co/chgk-files/Igry_dobroy_voli-2016.doc/Igry_dobroy_voli-2016.doc.html 1 ~~<s>\d+\. (.*?)Ответ~~~~<s>Ответ:(.*?)Автор:
	//!startregex http://kubgor2017.livejournal.com/905.html 1 ~~upictitle...kubgor2017: Корнелл...article...(.*?)Ответ\W~~~~Ответ\W(.*?)Автор
	cmd := `!startregex http://sp300p.livejournal.com/198595.html 1 ~~<div style='margin-left: 5px'>\d+. (.*?)<span class="lj-spoiler">~~~~<span class="lj-spoiler-body">Ответ:(.*?)<div class='entry-bottom-links'`
	s, _ := dbRegexp("http://sp300p.livejournal.com/198595.html", data, cmd)
	for _, q := range s.Questions {
		fmt.Println(">>>Question", q.Question)
		fmt.Println(">>>Picture", q.Picture)
		fmt.Println(">>>Answer", q.Answer)
	}
}
