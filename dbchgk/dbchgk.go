package dbchgk

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Kane-Sendgrid/chgk-bot/strip"
)

var picRE = regexp.MustCompile(`\(pic: (.*?)\)`)

var kandQRE = regexp.MustCompile(`(?s)<strong>Вопрос \d+:(.+?)<div style="display:none`)
var kandARE = regexp.MustCompile(`(?s)<strong>Ответ:<\/strong>(.+?)<a id=`)
var kandIMGRE = regexp.MustCompile(`(?s)<img src = "(.+?)"`)

var customQRE = regexp.MustCompile(`(?s)<p>(Вопрос)?\s*\d+\.(.*?)Ответ\b`)
var customAddRE = regexp.MustCompile(`(?s)Автор: .*?<\/p>(.+)`)
var customARE = regexp.MustCompile(`(?s)Ответ\b(.*?)Автор\b`)
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
func LoadSuite(url, gameType, cmd string) (*Suite, error) {
	fmt.Println("LoadSuite url:", url)
	if strings.Contains(url, "db.chgk.info") || strings.Contains(url, "pda.baza-voprosov.ru") {
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
	if gameType == "regex" {
		return dbRegexp(url, data, cmd)
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
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
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
		q.Question = strings.Replace(q.Question, "\n", " ", -1)
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
	matches := customQRE.FindAllStringSubmatch(str, -1)
	for _, m := range matches {
		images := customIMGRE.FindAllStringSubmatch(m[2], -1)
		img := ""
		for _, i := range images {
			img = url + i[1]
		}
		// fmt.Println(">>>IMAGE", img)
		// fmt.Println(m[2])
		q := m[2]
		addMatches := customAddRE.FindAllStringSubmatch(m[2], -1)
		if addMatches != nil && len(addMatches[0][1]) > 0 {
			q = addMatches[0][1]
		}
		s.Questions = append(s.Questions, &Question{
			Question: sanitize(q),
			Picture:  img,
		})
	}
	matches = customARE.FindAllStringSubmatch(str, -1)
	for i, m := range matches {
		s.Questions[i].Answer = sanitize(m[1])
	}
	return s, nil
}

func sanitize(s string) string {
	s = strings.Replace(s, "<br>", "\n", -1)
	s = strings.Replace(s, "<br/>", "\n", -1)
	s = strings.Replace(s, "<br />", "\n", -1)
	s = strings.Replace(s, "</p><p>", "\n", -1)
	s = strings.Replace(s, "</li><li>", "\n", -1)
	s = strip.StripTags(s)
	s = strings.Replace(s, "&nbsp;", " ", -1)
	s = strings.Replace(s, "&quot;", `"`, -1)
	s = html.UnescapeString(s)
	return s
}

func dbRegexp(url string, data []byte, cmd string) (*Suite, error) {
	cmd = html.UnescapeString(cmd)
	options := strings.Split(cmd, "~~")
	if len(options) != 4 {
		return nil, errors.New("Неправильное число аргументов: Пример !начать URL START_QUESTION_NUMBER~~QUESTION_REGEXP~~PIC_BASE_URL~~ANSWER_REGEXP")
	}
	fmt.Println(len(options))
	qRegexp := regexp.MustCompile(options[1])
	imgBaseURL := options[2]
	aRegexp := regexp.MustCompile(options[3])
	qPicRegexp := regexp.MustCompile(`(?s)<img.*?src\s?=\s?"(.+?)"`)

	s := &Suite{}
	str := string(data)
	matches := qRegexp.FindAllStringSubmatch(str, -1)
	for _, m := range matches {
		q := m[1]
		images := qPicRegexp.FindAllStringSubmatch(q, -1)
		img := ""
		for _, i := range images {
			img = imgBaseURL + i[1]
			q = img + "\n" + q
		}

		s.Questions = append(s.Questions, &Question{
			Question: sanitize(q),
			Picture:  img,
		})
	}
	matches = aRegexp.FindAllStringSubmatch(str, -1)
	for i, m := range matches {
		s.Questions[i].Answer = sanitize(m[1])
	}

	return s, nil
}
