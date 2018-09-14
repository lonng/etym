package memdb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"etym/cmd/etymd/protocol"
	"etym/cmd/etymd/syncer"
	"etym/pkg/db/model"
	"etym/pkg/log"
	"io/ioutil"
	"sort"
	"strings"
)

var (
	initOnce sync.Once

	// 索引器, 负责索引, 所有以单词作为key的地方, 都需要strings.ToLower处理
	indexes struct {
		etymology   map[string][]*model.Word                  // 所有词源(词干的其他派生词关联到词干的词源)
		dictionary  map[string]*protocol.ResDictWord          // 单词的翻译数据
		crossRefs   map[string][]*protocol.ResDictWord        // 交叉引用数据
		searchIndex map[byte]map[byte][]*protocol.ResDictWord // 搜索树索引, 目前索引树深度为2
	}
)

func init() {
	indexes.etymology = make(map[string][]*model.Word)
	indexes.dictionary = make(map[string]*protocol.ResDictWord)
	indexes.crossRefs = make(map[string][]*protocol.ResDictWord)
	indexes.searchIndex = make(map[byte]map[byte][]*protocol.ResDictWord)
}

func Load(dataDir string) {
	var load = func() {
		var err error
		dataDir, err = filepath.Abs(dataDir)
		if err != nil {
			panic(err)
		}

		log.Infof("Memdb load data from: %s", dataDir)

		var etymPath = filepath.Join(dataDir, "etym.json")
		var starPath = filepath.Join(dataDir, "stardict.json")
		var lammaPath = filepath.Join(dataDir, "lemma.en.txt")

		log.Infof("Etymology file path: %s", etymPath)
		log.Infof("Dictionary file path: %s", starPath)
		log.Infof("Lemma file path: %s", lammaPath)

		lemmaFile, err := ioutil.ReadFile(lammaPath)
		if err != nil {
			panic(err)
		}

		etymFile, err := os.OpenFile(etymPath, os.O_RDONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer etymFile.Close()

		starFile, err := os.OpenFile(starPath, os.O_RDONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer starFile.Close()

		// 加载词源和词典
		var etymWords []*model.Word
		var starWords []*protocol.ResDictWord
		if err := json.NewDecoder(etymFile).Decode(&etymWords); err != nil {
			panic(err)
		}

		if err := json.NewDecoder(starFile).Decode(&starWords); err != nil {
			panic(err)
		}

		log.Infof("Loading etymology completed, count: %d", len(etymWords))
		log.Infof("Loading dictionary completed, count: %d", len(starWords))

		// 将词源分组
		log.Info("Starting index etymology words")
		for _, etym := range etymWords {
			name := strings.ToLower(model.CleanWord(etym.Name))
			indexes.etymology[name] = append(indexes.etymology[name], etym)
		}
		log.Info("Indexing etymology words completed")
		log.Infof("Merge etymology lemma")

		// 拓展词干的派生词
		var mergeCount int
		lemmaList := strings.Split(string(lemmaFile), "\n")
		for _, line := range lemmaList {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, ";") {
				continue
			}
			parts := strings.Split(line, "->")
			if len(parts) != 2 {
				continue
				log.Warnf("Illegal lemma item: %s", line)
			}
			var lemma = strings.ToLower(parts[0])
			if ll := strings.Split(lemma, "/"); len(ll) > 1 {
				lemma = ll[0]
			}

			var derrives = strings.Split(parts[1], ",")
			for j, word := range derrives {
				derrives[j] = strings.ToLower(strings.TrimSpace(word))
			}

			lemmaEtym, found := indexes.etymology[lemma]
			if !found {
				//log.Warnf("Lemma not found in etymology, word: %s", lemma)
				continue
			}

			for _, word := range derrives {
				if _, found := indexes.etymology[word]; found {
					continue
				}
				mergeCount++
				indexes.etymology[word] = lemmaEtym
				//log.Infof("Merge lemma: %s, derive: %s", lemma, word)
			}
		}
		log.Infof("Merge etymology content: %d", mergeCount)

		// 索引单词翻译
		log.Info("Starting index dictionary translation")
		for _, word := range starWords {
			indexes.dictionary[strings.ToLower(word.Word)] = word
		}
		log.Info("Indexing dictionary translation completed")

		log.Info("Checking translation")
		var notfound int
		for word := range indexes.etymology {
			_, found := indexes.dictionary[word]
			if !found {
				notfound++
				//log.Warnf("Word: '%s' translation not found", word)
			}
		}
		log.Infof("Checking translation completed, count: %d", notfound)

		log.Infof("Starting index cross reference")
		for word, etyms := range indexes.etymology {
			var refs = map[string]struct{}{}
			for _, etym := range etyms {
				for _, ref := range etym.Ref {
					refs[strings.ToLower(ref)] = struct{}{}
				}
			}
			for ref := range refs {
				dict, found := indexes.dictionary[word]
				if !found {
					continue
				}
				indexes.crossRefs[ref] = append(indexes.crossRefs[ref], dict)
			}
		}
		log.Info("Cross reference indexing completed")

		// 构建搜索树(按照字母顺序排序)
		log.Infof("Start create search index")
		for word, dict := range indexes.dictionary {
			if len(word) < 1 {
				log.Warn("Empty word in dictionary")
				continue
			}

			var index1 = word[0]
			var index2 byte
			if len(word) == 1 {
				index2 = 0xFF
			} else {
				index2 = word[1]
			}
			level2, found := indexes.searchIndex[index1]
			if !found {
				level2 = make(map[byte][]*protocol.ResDictWord)
				indexes.searchIndex[index1] = level2
			}
			level2[index2] = append(level2[index2], dict)
		}
		for _, level1 := range indexes.searchIndex {
			for _, level2 := range level1 {
				sort.Slice(level2, func(i, j int) bool {
					wi := strings.TrimRight(level2[i].Word, ". ")
					wj := strings.TrimRight(level2[j].Word, ". ")
					return strings.Compare(wi, wj) < 0
				})
			}
		}
		log.Infof("Creating search index completed")
		log.Infof("Loading memdb data completed")
	}
	initOnce.Do(load)
}

func refRelated(ref string) []*protocol.ResDictWord {
	return indexes.crossRefs[ref]
}

func FindDictItem(word string) *protocol.ResDictWord {
	return indexes.dictionary[strings.ToLower(word)]
}

func HasEtymology(word string) bool {
	_, found := indexes.etymology[strings.ToLower(word)]
	return found
}

func HasDictItem(word string) bool {
	_, found := indexes.dictionary[strings.ToLower(word)]
	return found
}

func Etym(word string) *protocol.ResEtym {
	word = model.CleanWord(word)
	word = strings.ToLower(word)

	displayWord := word
	etymInfo := indexes.etymology[word]
	if len(etymInfo) > 0 {
		displayWord = model.CleanWord(etymInfo[0].Name)
	}

	var (
		resEtymInfo []*protocol.ResEtymInfo
		foreigns    []*protocol.ResDictWord
		references  []*protocol.Reference
	)

	if len(etymInfo) > 0 {
		resEtymInfo = make([]*protocol.ResEtymInfo, len(etymInfo))
		for i, etym := range etymInfo {
			etymCN, final := syncer.Find(etym.Name)
			resEtymInfo[i] = &protocol.ResEtymInfo{
				Name:    etym.Name,
				Etym:    etym.Etym,
				EtymCN:  etymCN,
				Final:   final,
				Ref:     etym.Ref,
				Foreign: etym.Foreign,
			}
		}

		var uniqueRef = map[string]struct{}{}
		var uniqueForeign = map[string]struct{}{}
		for _, etym := range etymInfo {
			for _, ref := range etym.Ref {
				uniqueRef[strings.ToLower(ref)] = struct{}{}
			}
		}

		for _, etym := range etymInfo {
			for _, foreign := range etym.Foreign {
				word := strings.ToLower(foreign)
				if _, found := uniqueRef[word]; found {
					continue
				}
				uniqueForeign[word] = struct{}{}
			}
		}

		for ref := range uniqueRef {
			references = append(references, &protocol.Reference{
				Word:    ref,
				Dict:    indexes.dictionary[ref],
				Related: refRelated(ref),
			})
		}

		for foreign := range uniqueForeign {
			d, found := indexes.dictionary[foreign]
			if !found {
				continue
			}
			foreigns = append(foreigns, d)
		}
	}

	etym := &protocol.ResEtym{
		Word:    displayWord,
		Etym:    resEtymInfo,
		Trans:   indexes.dictionary[word],
		Ref:     references,
		Related: foreigns,
	}
	return etym
}

func Search(word string) []*protocol.ResDictWord {
	word = strings.TrimSpace(word)
	if len(word) < 1 || word == "" {
		return []*protocol.ResDictWord{}
	}

	const searchLimit = 10
	level1, found := indexes.searchIndex[word[0]]
	if !found {
		return []*protocol.ResDictWord{}
	}

	// search in subtree
	if len(word) > 1 {
		level2, found := level1[word[1]]
		if !found {
			return []*protocol.ResDictWord{}
		}
		var result []*protocol.ResDictWord
		for _, dictItem := range level2 {
			if strings.HasPrefix(dictItem.Word, word) {
				result = append(result, dictItem)
			}
			if len(result) >= searchLimit {
				break
			}
		}
		return result
	}

	// search in all subtree
	var result []*protocol.ResDictWord
FindAll:
	for _, level2 := range level1 {
		for _, dictItem := range level2 {
			if strings.HasPrefix(dictItem.Word, word) {
				result = append(result, dictItem)
			}
			if len(result) >= searchLimit {
				break FindAll
			}
		}
	}
	return result
}
