package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"etym/pkg/db/model"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"etym/cmd/etymd/services/etym/memdb"
)

type WordInfo struct {
	Word     []*model.Word
	Trans    string `json:"trans"`
	Phonetic string `json:"phonetic"`
}

func init() {
	memdb.Load("/Users/lonng/devel/go/src/etym/assets", nil)
}

// 单词的翻译和词源, 词源包括词源原文和交叉引用
func generateJSON(wordlist []string, outpath string) {
	var outinfo []*WordInfo
	var skipnum int

	for _, word := range wordlist {
		if len(word) < 1 {
			continue
		}
		etym := memdb.FindEtymology(word)
		dict := memdb.FindDictItem(word)
		if etym == nil || len(etym) == 0 || dict == nil {
			skipnum += 1
			fmt.Printf("%4d Skip: %s\n", skipnum, word)
			continue
		}
		outinfo = append(outinfo, &WordInfo{
			Word:     etym,
			Trans:    dict.Translation,
			Phonetic: dict.Phonetic,
		})
	}
	var outbuf = bytes.NewBuffer(nil)
	json.NewEncoder(outbuf).Encode(outinfo)
	var dstbuf = bytes.NewBuffer(nil)
	json.Indent(dstbuf, outbuf.Bytes(), "", "\t")

	if err := ioutil.WriteFile(outpath, dstbuf.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
}

// 单词的翻译和词源, 词源包括词源原文和交叉引用
func generateIndex(wordlist []string, outpath string) {
	var words = map[string]struct{}{}
	for _, word := range wordlist {
		words[word] = struct{}{}
	}

	type RefInfo struct {
		Trans    string
		Phonetic string
		Word     string
		Ref      map[string]struct{}
		Foreign  map[string]struct{}
		Index    int
	}

	var wordsinfo = map[string]*RefInfo{}
	for _, word := range wordlist {
		if len(word) < 1 || /* skip phrase */ strings.Index(word, " ") >= 0 {
			continue
		}
		etym := memdb.FindEtymology(word)
		dict := memdb.FindDictItem(word)
		if /*etym == nil || len(etym) == || 0*/ dict == nil {
			continue
		}

		var uniqueRef = map[string]struct{}{}
		var uniqueFor = map[string]struct{}{}
		for _, e := range etym {
			for _, ref := range e.Ref {
				uniqueRef[ref] = struct{}{}
			}
			for _, ref := range e.Foreign {
				uniqueFor[ref] = struct{}{}
			}
		}
		info := &RefInfo{
			Word:     word,
			Trans:    strings.Replace(dict.Translation, "\\n", "; ", -1),
			Phonetic: dict.Phonetic,
			Ref:      uniqueRef,
			Foreign:  uniqueFor,
		}
		wordsinfo[word] = info
	}

	fmt.Println("wordsinfo", len(wordsinfo))
	var refsinfo = map[string][]string{}
	var sortWords []string
	for _, word := range wordsinfo {
		sortWords = append(sortWords, word.Word)
		var ref = word.Ref
		for r := range ref {
			refsinfo[r] = append(refsinfo[r], word.Word)
		}
	}

	sort.Strings(sortWords)
	for i, w := range sortWords {
		wordsinfo[w].Index = i
	}

	type Item struct {
		Word     string `json:"word"`
		Phonetic string `json:"phonetic"`
		Trans    string `json:"trans"`
		Derives  []int  `json:"derives,omitemtpy"`
		Audio    string `json:"-"`
	}

	var results []*Item
	for _, w := range sortWords {
		word := wordsinfo[w]
		refWords := map[string]struct{}{}
		for foreign := range word.Foreign {
			refWords[foreign] = struct{}{}
		}
		for ref := range word.Ref {
			refInfo := refsinfo[ref]
			for _, w := range refInfo {
				refWords[w] = struct{}{}
			}
		}

		var derives []int
		for ref := range refWords {
			if ref == word.Word {
				continue
			}
			inner := wordsinfo[ref]
			if inner == nil {
				continue
			}
			derives = append(derives, inner.Index)
		}
		var audio string
		audioPath := filepath.Join("./audio", fmt.Sprintf("%s.ogg", word.Word))
		if content, err := ioutil.ReadFile(audioPath); err == nil {
			audio = base64.StdEncoding.EncodeToString(content)
		}
		results = append(results, &Item{
			Word:     word.Word,
			Phonetic: word.Phonetic,
			Trans:    word.Trans,
			Derives:  derives,
			Audio:    audio,
		})
	}

	var outbuf = bytes.NewBuffer(nil)
	json.NewEncoder(outbuf).Encode(results)
	// var dstbuf = bytes.NewBuffer(nil)
	//json.Indent(dstbuf, outbuf.Bytes(), "", "\t")

	if err := ioutil.WriteFile(outpath, outbuf.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
}

// 单词的翻译和词源, 词源包括词源原文和交叉引用
func generateSimpleMD(wordlist []string, outpath string) {
	var words = map[string]struct{}{}
	for _, word := range wordlist {
		words[word] = struct{}{}
	}

	type RefInfo struct {
		Trans    string
		Phonetic string
		Word     string
		Ref      map[string]struct{}
		Foreign  map[string]struct{}
	}

	var wordsinfo = map[string]*RefInfo{}
	for _, word := range wordlist {
		if len(word) < 1 || /* skip phrase */ strings.Index(word, " ") >= 0 {
			continue
		}
		etym := memdb.FindEtymology(word)
		dict := memdb.FindDictItem(word)
		if /*etym == nil || len(etym) == || 0*/ dict == nil {
			continue
		}

		var uniqueRef = map[string]struct{}{}
		var uniqueFor = map[string]struct{}{}
		for _, e := range etym {
			for _, ref := range e.Ref {
				uniqueRef[ref] = struct{}{}
			}
			for _, ref := range e.Foreign {
				uniqueFor[ref] = struct{}{}
			}
		}
		info := &RefInfo{
			Word:     word,
			Trans:    dict.Translation,
			Phonetic: dict.Phonetic,
			Ref:      uniqueRef,
			//Foreign:  uniqueFor,
		}
		wordsinfo[word] = info
	}

	fmt.Println("wordsinfo", len(wordsinfo))
	var refsinfo = map[string][]string{}
	for _, word := range wordsinfo {
		var ref = word.Ref
		for r := range ref {
			refsinfo[r] = append(refsinfo[r], word.Word)
		}
	}

	out := &bytes.Buffer{}
	for _, word := range wordsinfo {
		refWords := map[string]struct{}{}
		for foreign := range word.Foreign {
			refWords[foreign] = struct{}{}
		}
		for ref := range word.Ref {
			refInfo := refsinfo[ref]
			for _, w := range refInfo {
				refWords[w] = struct{}{}
			}
		}

		fmt.Fprintf(out, "### %s `/%s/ %s`\n", word.Word, word.Phonetic, strings.Replace(word.Trans, "\\n", "; ", -1))
		for ref := range refWords {
			if ref == word.Word {
				continue
			}
			inner := wordsinfo[ref]
			if inner == nil {
				continue
			}
			fmt.Fprintf(out, "- %s `/%s/ %s`\n", inner.Word, inner.Phonetic, strings.Replace(inner.Trans, "\\n", "; ", -1))
		}
	}

	if err := ioutil.WriteFile(outpath, out.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
}

// 单词的翻译和词源, 词源包括词源原文和交叉引用
func generateMD(wordlist []string, outpath string) {
	var words = map[string]struct{}{}
	for _, word := range wordlist {
		words[word] = struct{}{}
	}

	type RefInfo struct {
		Trans    string
		Phonetic string
		Word     string
		Ref      map[string]struct{}
	}

	var wordsinfo = map[string]*RefInfo{}
	for _, word := range wordlist {
		if len(word) < 1 {
			continue
		}
		etym := memdb.FindEtymology(word)
		dict := memdb.FindDictItem(word)
		if /*etym == nil || len(etym) == || 0*/ dict == nil {
			continue
		}

		var uniqueRef = map[string]struct{}{}
		for _, e := range etym {
			for _, ref := range e.Ref {
				uniqueRef[ref] = struct{}{}
			}
		}
		info := &RefInfo{
			Word:     word,
			Trans:    dict.Translation,
			Phonetic: dict.Phonetic,
			Ref:      uniqueRef,
		}
		wordsinfo[word] = info
	}

	fmt.Println("wordsinfo", len(wordsinfo))
	var refsinfo = map[string][]*RefInfo{}
	for _, word := range wordsinfo {
		var ref = word.Ref
		for r := range ref {
			refsinfo[r] = append(refsinfo[r], word)
		}
	}

	fmt.Println("refsinfo", len(refsinfo))
	var refs []*struct {
		w string
		c int
	}
	for ref, ws := range refsinfo {
		refs = append(refs, &struct {
			w string
			c int
		}{ref, len(ws)})
	}

	out := &bytes.Buffer{}
	sort.Slice(refs, func(i, j int) bool { return refs[i].c > refs[j].c })
	for _, ref := range refs {
		words := refsinfo[ref.w]
		if len(words) == 1 {
			continue
		}
		var trans string
		if d := memdb.FindDictItem(ref.w); d != nil {
			trans = d.Translation
		}
		fmt.Fprintf(out, "### %s(%d)\n", ref.w, ref.c)
		if trans == "" {
			var etym = memdb.FindEtymology(ref.w)
			for _, e := range etym {
				fmt.Fprintf(out, "%s\n", e.Etym)
			}
		} else {
			fmt.Fprintf(out, "%s\n", strings.Replace(trans, "\\n", "; ", -1))
		}
		for _, w := range words {
			if len(w.Phonetic) > 0 {
				fmt.Fprintf(out, "- ***%s*** `/%s/` %s\n", w.Word, w.Phonetic, strings.Replace(w.Trans, "\\n", "; ", -1))
			} else {
				fmt.Fprintf(out, "- ***%s*** %s\n", w.Word, strings.Replace(w.Trans, "\\n", "; ", -1))
			}
		}
	}

	fmt.Fprintf(out, "---\n")
	for _, ref := range refs {
		words := refsinfo[ref.w]
		if len(words) == 1 {
			for _, w := range words {
				if len(w.Phonetic) > 0 {
					fmt.Fprintf(out, "- ***%s*** `/%s/` %s\n", w.Word, w.Phonetic, strings.Replace(w.Trans, "\\n", "; ", -1))
				} else {
					fmt.Fprintf(out, "- ***%s*** %s\n", w.Word, strings.Replace(w.Trans, "\\n", "; ", -1))
				}
			}
		}
	}

	if err := ioutil.WriteFile(outpath, out.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
}

func downloadAudio(wordlist []string, outpath string) {
	var words = map[string]struct{}{}
	for _, word := range wordlist {
		if len(word) < 1 || /* skip phrase */ strings.Index(word, " ") >= 0 {
			continue
		}
		dict := memdb.FindDictItem(word)
		if /*etym == nil || len(etym) == || 0*/ dict == nil {
			continue
		}
		words[word] = struct{}{}
	}

	var succ int32

	ch := make(chan string, 100)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for word := range ch {
				path := filepath.Join(outpath, fmt.Sprintf("%s.ogg", word))
				if _, err := os.Stat(path); err == nil {
					atomic.AddInt32(&succ, 1)
					continue
				}
				url := fmt.Sprintf("http://dict.youdao.com/dictvoice?audio=%s&type=2", word)
				fmt.Println("downloading", url)
				resp, err := http.Get(url)
				if err != nil {
					continue
				}
				bytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					continue
				}
				if err := ioutil.WriteFile(path, bytes, os.ModePerm); err != nil {
					continue
				}
				atomic.AddInt32(&succ, 1)
			}
		}()
	}

	for word := range words {
		ch <- word
	}
	close(ch)
	wg.Wait()
	fmt.Println("Total", len(words), "Succ", succ)
}

func main() {
	content, err := ioutil.ReadFile("/Users/lonng/devel/opensource/vocabulary/unique.txt")
	if err != nil {
		panic(err)
	}
	words := strings.Split(string(content), "\n")
	for i := range words {
		words[i] = strings.ToLower(strings.TrimSpace(words[i]))
	}
	// generateJSON(words, "./unique.json")
	// generateMD(words, "./unique.md")
	generateIndex(words, "./unique.index")
	//downloadAudio(words, "./audio")
}
