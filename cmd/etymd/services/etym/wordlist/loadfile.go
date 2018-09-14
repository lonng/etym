package wordlist

import (
	"etym/cmd/etymd/services/etym/memdb"
	"etym/pkg/log"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

func loadfile(dataDir string, value int, file string) {
	var loadFile = filepath.Join(dataDir, file)
	log.Infof("Load word list file path: %s", loadFile)
	cocaFile, err := ioutil.ReadFile(loadFile)
	if err != nil {
		panic(err)
	}

	cocaList := strings.Split(string(cocaFile), "\n")
	var dictnotfound, etymnotfound int
	for i, word := range cocaList {
		word = strings.ToLower(strings.TrimSpace(word))
		if !memdb.HasDictItem(word) {
			dictnotfound++
		}

		if !memdb.HasEtymology(word) {
			etymnotfound++
		}

		cocaList[i] = word
	}

	registered[strconv.Itoa(value)] = cocaList
	log.Infof("Loading word list: dictionary not found: %d, etymology not found: %d, path: %s",
		dictnotfound, etymnotfound, loadFile)
}
