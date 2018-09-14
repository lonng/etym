package model

import "strings"

type (
	Sentence struct {
		Trans string `json:"trans"`
		Orig  string `json:"orig"`
	}

	// 保存完整翻译, 不按句保存
	PureTranslate struct {
		Word  string `json:"word"`
		Trans string `json:"trans"`
		Final bool   `json:"final"`
	}

	SpiderTranslate struct {
		Sentences []*Sentence `json:"sentences"`
	}

	// 按句保存
	Translate struct {
		Word  string           `json:"word"`
		Trans *SpiderTranslate `json:"trans"`
	}

	Word struct {
		Name    string   `json:"name"`
		Etym    string   `json:"etym"`
		Ref     []string `json:"ref"`
		Foreign []string `json:"foreign"`
	}

	DictWord struct {
		Word        string `json:"word"`
		Phonetic    string `json:"phonetic"`
		Translation string `json:"translation"`
	}
)

func CleanWord(word string) string {
	if index := strings.Index(word, "("); index > 0 {
		word = word[:index]
	}
	return strings.TrimSpace(word)
}
