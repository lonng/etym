package command

import (
	"encoding/json"
	"etym/pkg/constant"
	"etym/pkg/db/model"
	"etym/pkg/log"
	"html"
	"os"
	"path/filepath"
	"strings"
)

func CombineTrans() {
	var homeDir = "../../../etymology-resource/etymology"
	var err error
	homeDir, err = filepath.Abs(homeDir)
	if err != nil {
		panic(err)
	}
	var wordsDir = filepath.Join(homeDir, constant.TranslationVer)

	log.Infof("Starting load translation from: %s", homeDir)
	log.Infof("Words directory: %s", wordsDir)

	var totalCount int
	filepath.Walk(wordsDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || err != nil {
			return nil
		}
		totalCount++
		return nil
	})

	var words []*model.PureTranslate

	outfile, err := os.OpenFile("../../assets/trans.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	var currentCount int
	filepath.Walk(wordsDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || err != nil {
			return nil
		}
		wordfile, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Warnf("Open file failed, error: %v", err)
			return nil
		}
		defer wordfile.Close()

		etym := new(model.Translate)
		if err := json.NewDecoder(wordfile).Decode(etym); err != nil {
			log.Warnf("Decode file:%s, error: %v", path, err)
			return nil
		}

		if etym.Trans == nil {
			log.Infof("Empty trans, path:%s", path)
			return nil
		}

		etym.Word = strings.TrimSpace(html.UnescapeString(etym.Word))
		word := &model.PureTranslate{}
		word.Word = etym.Word
		result := combineSentences(etym.Trans.Sentences)
		word.Trans = result.Translation
		word.Final = result.Final > 0
		words = append(words, word)
		currentCount++
		if currentCount%1000 == 0 {
			log.Infof("Loading progress, current word count: %5d / %5d", currentCount, totalCount)
		}

		return nil
	})

	if err := json.NewEncoder(outfile).Encode(words); err != nil {
		log.Errorf("Write combine file failed: %v", err)
	}

	log.Infof("Combine translation count: %d/%d", currentCount, totalCount)
}
