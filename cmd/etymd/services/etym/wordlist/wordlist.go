package wordlist

import (
	"etym/cmd/etymd/protocol"
	"etym/cmd/etymd/services/etym/memdb"
	"etym/pkg/errutil"
	"etym/pkg/log"
	"github.com/lonnng/nex"
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

var (
	loadOnce sync.Once

	registered = map[string][]string{}
)

func Load(dataDir string) {
	var load = func() {
		var files = []struct {
			v int
			f string
		}{
			{wordListTagChuZhong, "chuzhong2182.txt"},
			{wordListTagGaoZhong, "gaozhong3500.txt"},
			{wordListTagZSB, "zsb.txt"},
			{wordListTagCET4, "CET4.txt"},
			{wordListTagCET6, "CET6.txt"},
			{wordListTagCoca, "coca20000.txt"},
		}

		for _, file := range files {
			loadfile(dataDir, file.v, file.f)
		}

		log.Infof("Load word list completed")
		for key, list := range registered {
			log.Infof("Word list id: %s, count: %d", key, len(list))
		}
	}
	loadOnce.Do(load)
}

func Rand(query *nex.Form) (*protocol.ResEtym, error) {
	// 默认使用初高中
	var rang = map[string]struct{}{
		strconv.Itoa(wordListTagChuZhong): {},
		strconv.Itoa(wordListTagGaoZhong): {},
	}

	liststr := strings.TrimSpace(query.Get("range"))
	if list := strings.Split(liststr, ","); len(list) > 0 {
		cuzRang := map[string]struct{}{}
		for _, s := range list {
			key := strings.TrimSpace(s)
			if d, found := registered[key]; found && len(d) > 0 {
				cuzRang[key] = struct{}{}
			}
		}
		if len(cuzRang) > 0 {
			rang = cuzRang
		}
	}

	for key := range rang {
		dict := registered[key]
		leng := len(dict)
		if leng < 1 {
			continue
		}
		var limit = 0
		for limit < 100 {
			limit++
			word := dict[rand.Intn(leng)]
			if !memdb.HasEtymology(word) {
				continue
			}
			return memdb.Etym(word), nil
		}
	}

	return nil, errutil.ErrServiceUnavailable
}

const (
	wordListTagBegin    = 0
	wordListTagChuZhong = 1  // 初中
	wordListTagGaoZhong = 2  // 高中
	wordListTagZSB      = 3  // 专升本
	wordListTagCET4     = 4  // 大学四级
	wordListTagCET6     = 5  // 大学六级
	wordListTagKaoYan   = 6  // 考研
	wordListTagKaoBo    = 7  // 考博
	wordListTagITETS    = 8  // 雅思
	wordListTagTofl     = 9  // 托福
	wordListTagSAT      = 10 // SAT
	wordListTagGRE      = 11 // GRE
	wordListTagCoca     = 12 // COCA20000
	wordListTagEnd      = 13
)

var supportedWordlist = &protocol.ResWordList{
	Items: []*protocol.WordListItem{
		{wordListTagChuZhong, "初中"},
		{wordListTagGaoZhong, "高中"},
		{wordListTagZSB, "专升本"},
		{wordListTagCET4, "CET4"},
		{wordListTagCET6, "CET6"},
		//{wordListTagKaoYan, "考研"},
		//{wordListTagKaoBo, "考博"},
		//{wordListTagITETS, "雅思"},
		//{wordListTagTofl, "托福"},
		//{wordListTagSAT, "SAT"},
		//{wordListTagGRE, "GRE"},
		{wordListTagCoca, "COCA"},
	},
}

func WordList() (*protocol.ResWordList, error) {
	return supportedWordlist, nil
}
