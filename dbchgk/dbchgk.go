package dbchgk

import (
	"encoding/xml"
	"io/ioutil"
	"regexp"
)

var picRE = regexp.MustCompile(`\(pic: (.*)\)`)

//Question ...
type Question struct {
	Question     string
	Answer       string
	PassCriteria string
	Authors      string
	Sources      string
	Comments     string
	Picture      string
}

//Suite ...
type Suite struct {
	XMLName   xml.Name    `xml:"tournament"`
	Questions []*Question `xml:"question"`
}

//LoadSuite ...
func LoadSuite() *Suite {
	s := &Suite{}
	data, _ := ioutil.ReadFile("/Users/kane/work/code/sendgrid/github/go/src/github.com/Kane-Sendgrid/chgk-bot/testdata/chgk.info.xml")
	xml.Unmarshal(data, s)
	for _, q := range s.Questions {
		matches := picRE.FindAllStringSubmatch(q.Question, -1)
		for _, m := range matches {
			q.Picture = "http://db.chgk.info/images/db/" + m[1]
		}
	}
	return s
}
