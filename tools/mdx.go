package main

import (
	"bytes"
	"encoding/json"
	"etym/pkg/db/model"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type xmlWord struct {
}

func main() {
	starFile, err := os.Open("assets/stardict.json")
	if err != nil {
		panic(err)
	}

	var starWords []*model.DictWord
	if err := json.NewDecoder(starFile).Decode(&starWords); err != nil {
		panic(err)
	}
	starFile.Close()

	println("Stardict words =>", len(starWords))

	data, err := ioutil.ReadFile(`D:\Desktop\xwang-mdict-analysis-b17ca266d70c\tmp\21世纪大英汉词典.html`)
	if err != nil {
		panic(err)
	}

	phoneRegexp := regexp.MustCompile(`<span class="phone">(.*?)</span>`)

	var parsedWords []*model.DictWord
	words := strings.Split(string(data), "=========================||===========================")
	//words := strings.Split(string(data), `<link rel="stylesheet" type="text/css" href="sf_ecce.css"/>`)
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		parts := strings.Split(w, `<link rel="stylesheet" type="text/css" href="sf_ecce.css"/>`)
		if len(parts) != 2 {
			continue
		}

		phones := phoneRegexp.FindStringSubmatch(parts[1])
		var phone string
		if len(phones) == 2 && phones[1] != "" {
			phone = phones[1]
		}
		parsed := &model.DictWord{
			Word:     strings.TrimSpace(parts[0]),
			Phonetic: phone,
		}

		var doc, err = goquery.NewDocumentFromReader(bytes.NewBufferString(parts[1]))
		if err != nil {
			panic(err)
		}

		var translation []string
		doc.Find(".trs").Each(func(i int, selection *goquery.Selection) {
			var ss []string
			pos := selection.Find(".pos")
			posStr := strings.TrimSpace(pos.Text())
			if posStr != "" {
				ss = append(ss, posStr)
			}
			var tr []string
			selection.Find(".tr").Children().Each(func(j int, trans *goquery.Selection) {
				if s, found := trans.Attr("class"); found && s == "exam" {
					return
				}
				trans.Find(".l .i").Each(func(_ int, tran *goquery.Selection) {
					tr = append(tr, tran.Text())
				})
			})
			if len(tr) > 0 {
				if posStr == "suf." {
					ss = append(ss, strings.Join(tr, "\\n"))
				} else {
					ss = append(ss, strings.Join(tr, "；"))
				}
			}

			if len(ss) > 0 {
				if posStr == "suf." {
					translation = append(translation, strings.Join(ss, "\\n"))
				} else {
					translation = append(translation, strings.Join(ss, " "))
				}
			}
		})

		parsed.Translation = strings.Join(translation, "\\n")
		parsedWords = append(parsedWords, parsed)
	}

	starIndex := map[string]*model.DictWord{}
	for _, w := range starWords {
		if strings.Index(w.Word, " ") >= 0 {
			continue
		}
		isLetter := true
		for _, char := range w.Word {
			if char < 'A' || char > 'z' {
				isLetter = false
				break
			}
		}
		if isLetter {
			continue
		}
		starIndex[w.Word] = w
	}
	println("StarIndex =>", len(starIndex))

	toneIndex := map[string]*model.DictWord{}
	for _, w := range parsedWords {
		toneIndex[w.Word] = w
	}

	// merge
	mergeCount := 0
	for word, w := range starIndex {
		t21, found := toneIndex[word]
		if !found {
			continue
		}

		if t21.Translation == w.Translation {
			continue
		}
		//println("21===>", t21.Translation)
		//println(colorized.Blue("ST")+"===>", w.Translation)
		//println()
		if len(t21.Translation) > len(w.Translation) {
			w.Translation = t21.Translation
			mergeCount++
		}
	}

	println("merge count =>", mergeCount)

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(starWords); err != nil {
		panic(err)
	}

	dst := bytes.NewBuffer(nil)
	if err := json.Indent(dst, buf.Bytes(), "", "\t"); err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("assets/stardict.json", dst.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
}
