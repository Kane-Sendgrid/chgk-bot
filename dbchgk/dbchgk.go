package dbchgk

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Kane-Sendgrid/chgk-bot/strip"
)

var picRE = regexp.MustCompile(`\(pic: (.*)\)`)

var kandQRE = regexp.MustCompile(`(?s)<strong>Вопрос \d+:(.+?)<div style="display:none`)
var kandARE = regexp.MustCompile(`(?s)<strong>Ответ:<\/strong>(.+?)<a id=`)
var kandIMGRE = regexp.MustCompile(`(?s)<img src = "(.+?)"`)

var customQRE = regexp.MustCompile(`(?s)<p>(Вопрос)?\s*\d+\.(.*?)Ответ:`)
var customARE = regexp.MustCompile(`(?s)Ответ:(.*?)Автор`)
var customIMGRE = regexp.MustCompile(`(?s)<img.*?src="(.*?)"`)

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
func LoadSuite(url string) (*Suite, error) {
	fmt.Println("LoadSuite url:", url)
	if strings.Contains(url, "db.chgk.info") {
		data, err := loadData(url + "/xml")
		if err != nil {
			return nil, err
		}
		return dbCHGK(data)
	}
	data, err := loadData(url)
	if err != nil {
		return nil, err
	}

	if strings.Contains(url, "kand.info") {
		return dbKAND(data)
	}
	if strings.Contains(url, "kane1.ipq.co") {
		return dbCustom(url, data)
	}
	return nil, fmt.Errorf("Url not found")
}

func loadData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func dbCHGK(data []byte) (*Suite, error) {
	s := &Suite{}
	xml.Unmarshal(data, s)
	for _, q := range s.Questions {
		matches := picRE.FindAllStringSubmatch(q.Question, -1)
		for _, m := range matches {
			q.Picture = "http://db.chgk.info/images/db/" + m[1]
		}
	}
	return s, nil
}

func dbKAND(data []byte) (*Suite, error) {
	s := &Suite{}
	str := string(data)
	matches := kandQRE.FindAllStringSubmatch(str, -1)
	for _, m := range matches {
		images := kandIMGRE.FindAllStringSubmatch(m[1], -1)
		img := ""
		for _, i := range images {
			img = "http://kand.info" + i[1]
		}
		// fmt.Println(">>>IMAGE", img)
		s.Questions = append(s.Questions, &Question{
			Question: sanitize(m[1]),
			Picture:  img,
		})
	}
	matches = kandARE.FindAllStringSubmatch(str, -1)
	for i, m := range matches {
		s.Questions[i].Answer = sanitize(m[1])
	}
	return s, nil
}

func dbCustom(url string, data []byte) (*Suite, error) {
	fmt.Println("Loading custom file")
	url = url[0 : strings.LastIndex(url, "/")+1]
	s := &Suite{}
	str := string(data)
	fmt.Println("data", str)
	matches := customQRE.FindAllStringSubmatch(str, -1)
	fmt.Println("question matches", matches)
	for _, m := range matches {
		images := customIMGRE.FindAllStringSubmatch(m[2], -1)
		img := ""
		for _, i := range images {
			img = url + i[1]
		}
		// fmt.Println(">>>IMAGE", img)
		s.Questions = append(s.Questions, &Question{
			Question: sanitize(m[2]),
			Picture:  img,
		})
	}
	matches = customARE.FindAllStringSubmatch(str, -1)
	fmt.Println("answer matches", matches)
	for i, m := range matches {
		s.Questions[i].Answer = sanitize(m[1])
	}
	return s, nil
}

func sanitize(s string) string {
	s = strip.StripTags(s)
	s = strings.Replace(s, "&nbsp;", " ", -1)
	return s
}
