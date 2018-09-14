package command

import (
	"encoding/json"
	"etym/pkg/constant"
	"etym/pkg/db"
	"etym/pkg/db/model"
	"etym/pkg/log"
	"html"
	"os"
	"path/filepath"
	"strings"
)

func DBImport(args []string) {
	if len(args) < 1 {
		panic("请指明导入的类型")
	}

	db.Initialize("root:43OYHrP9Lc2SBeJxwpPA@tcp(172.16.64.135:3306)/etym?charset=utf8")
	//db.Initialize("root:password@tcp(127.0.0.1:3306)/etym?charset=utf8")

	switch args[0] {
	case "trans":
		importTrans()

	default:
		panic("支持的数据库类型" + args[0])
	}
}

func importTrans() {
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

	var database = db.Database()
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
		currentCount++
		if currentCount%1000 == 0 {
			log.Infof("Loading progress, current word count: %5d / %5d", currentCount, totalCount)
		}

		combined := combineSentences(etym.Trans.Sentences)
		record := model.DbHistory{
			Word:      etym.Word,
			Status:    model.TranslationStatusApproved,
			LowerWord: strings.ToLower(etym.Word),
			Original:  combined.Original,
			NextTrans: combined.Translation,
		}

		if combined.Final > 0 {
			record.Status = model.TranslationStatusFinal
		}

		/*_ = record
		_ = database*/
		if _, err := database.Insert(record); err != nil {
			log.Warnf("Insert record failure, word: %s, %v", etym.Word, err)
		}

		return nil
	})

	log.Infof("Combine translation count: %d/%d", currentCount, totalCount)
}
