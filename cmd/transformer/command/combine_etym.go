package command

import (
	"encoding/json"
	spider "etym/cmd/spider/model"
	"etym/pkg/db/model"
	"etym/pkg/log"
	"html"
	"os"
	"path/filepath"
	"strings"
)

func CombineEtym() {
	var homeDir = "../../etymology-resource/etymology"
	var err error
	homeDir, err = filepath.Abs(homeDir)
	if err != nil {
		panic(err)
	}
	var wordsDir = filepath.Join(homeDir, "export")

	log.Infof("Starting load etyms from: %s", homeDir)
	log.Infof("Words directory: %s", wordsDir)

	var totalCount int
	filepath.Walk(wordsDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || err != nil {
			return nil
		}
		totalCount++
		return nil
	})

	var words []*model.Word

	outfile, err := os.OpenFile("etym.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}

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

		etym := new(spider.Etym)
		if err := json.NewDecoder(wordfile).Decode(etym); err != nil {
			log.Warnf("Decode file:%s, error: %v", path, err)
			return nil
		}

		etym.Word = strings.TrimSpace(html.UnescapeString(etym.Word))
		for i := range etym.Ref {
			etym.Ref[i] = strings.TrimSpace(html.UnescapeString(etym.Ref[i]))
		}
		for i := range etym.Foreign {
			etym.Foreign[i] = strings.TrimSpace(html.UnescapeString(etym.Foreign[i]))
		}

		word := &model.Word{}
		word.Name = etym.Word
		word.Etym = etym.Etym
		word.Ref = etym.Ref
		word.Foreign = etym.Foreign

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

	log.Infof("Loaded etyms count: %d", currentCount)
}
