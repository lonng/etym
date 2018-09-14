package command

import (
	"encoding/json"
	"etym/pkg/db/model"
	"etym/pkg/log"
	"os"
)

func Merge() {
	starfile, err := os.OpenFile("stardict.json", os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer starfile.Close()

	ecfile, err := os.OpenFile("ecdict.json", os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer ecfile.Close()

	log.Infof("Compare file stardict.json and ecdict.json")

	var starwords, ecwords []*model.DictWord
	if err := json.NewDecoder(starfile).Decode(&starwords); err != nil {
		panic(err)
	}

	if err := json.NewDecoder(ecfile).Decode(&ecwords); err != nil {
		panic(err)
	}

	log.Infof("Found stardict words count: %d", len(starwords))
	log.Infof("Found ecdict words count: %d", len(ecwords))

	var starindex = map[string]*model.DictWord{}
	for _, word := range starwords {
		starindex[word.Word] = word
	}

	var merge []*model.DictWord
	// compare
	for _, word := range ecwords {
		starword, found := starindex[word.Word]
		if !found {
			//log.Warnf("Word found in ECDICT, but not in StarDICT: %s", word.Word)
			merge = append(merge, word)
			continue
		}
		if starword.Translation != word.Translation {
			//log.Infof("Found diff: Word=%s, ECTrans=%s, StarTrans=%s", word.Word, word.Translation, starword.Translation)
		}
	}

	log.Infof("Merge count: %d (Word found in ECDICT, but not in StarDICT)", len(merge))
	for _, word := range merge {
		starwords = append(starwords, word)
	}

	log.Infof("Will override stardict.json")
	starfile.Seek(0, 0)
	starfile.Truncate(0)
	if err := json.NewEncoder(starfile).Encode(starwords); err != nil {
		panic(err)
	}
	starfile.Sync()
	log.Infof("Merge complete")
}
