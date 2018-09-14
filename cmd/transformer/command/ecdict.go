package command

import (
	"encoding/csv"
	"encoding/json"
	"etym/pkg/db/model"
	"etym/pkg/log"
	"fmt"
	"os"
	"path/filepath"
)

func ECDICT() {
	var create = func(header, line []string) *model.DictWord {
		if len(header) != len(line) {
			panic(fmt.Errorf("illegal line, header:%+v, line:%+v", header, line))
		}
		var ret = &model.DictWord{}
		for i, title := range header {
			switch title {
			case "word":
				ret.Word = line[i]

			case "phonetic":
				ret.Phonetic = line[i]

			case "translation":
				ret.Translation = line[i]
			}
		}
		return ret
	}

	// /Users/lonng/devs/go/src/etym/assets/ECDICT/ecdict.csv
	ecdict := "../../assets/ECDICT/ecdict.csv"
	path, err := filepath.Abs(ecdict)
	if err != nil {
		panic(err)
	}

	log.Infof("Export ECDICT data, path: %s", path)

	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	outfile, err := os.OpenFile("ecdict.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		panic(err)
	}

	if len(lines) < 1 {
		log.Infof("empty csv file")
		os.Exit(0)
	}

	var words []*model.DictWord
	header := lines[0]
	for i, line := range lines {
		if i < 1 { // skip header
			continue
		}
		words = append(words, create(header, line))
	}

	if err := json.NewEncoder(outfile).Encode(words); err != nil {
		panic(err)
	}
	log.Infof("Export completed, words count: %d", len(words))
}
