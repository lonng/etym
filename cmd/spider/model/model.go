package model

import "etym/pkg/db/model"

type (
	Word struct {
		Word  string                 `json:"word"`
		Trans *model.SpiderTranslate `json:"trans"`
	}

	Etym struct {
		Word    string   `json:"word"`
		RawEtym string   `json:"raw_etym"`
		Etym    string   `json:"etym"`
		Ref     []string `json:"ref"`
		Foreign []string `json:"foreign"`
	}
)
