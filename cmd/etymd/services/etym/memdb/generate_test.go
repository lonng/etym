package memdb

import (
	"bytes"
	"encoding/json"
	"etym/pkg/db/model"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"testing"
)

type WordInfo struct {
	Word     []*model.Word
	Trans    string `json:"trans"`
	Phonetic string `json:"phonetic"`
}

func init() {
	Load("C:/devs/go/src/etym/assets")
}

// 单词的翻译和词源, 词源包括词源原文和交叉引用
func generateJSON(wordlist []string, outpath string) {
	out, err := os.OpenFile(outpath, os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	var outinfo []*WordInfo
	var skipnum int

	for _, word := range wordlist {
		if len(word) < 1 {
			continue
		}
		etym := indexes.etymology[word]
		dict := indexes.dictionary[word]
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
	io.Copy(out, dstbuf)
}

// 单词的翻译和词源, 词源包括词源原文和交叉引用
func generateMD(wordlist []string, outpath string) {
	out, err := os.OpenFile(outpath, os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer out.Close()

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
		etym := indexes.etymology[word]
		dict := indexes.dictionary[word]
		if etym == nil || len(etym) == 0 || dict == nil {
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

	var refsinfo = map[string][]*RefInfo{}
	for _, word := range wordsinfo {
		var ref = word.Ref
		for r := range ref {
			refsinfo[r] = append(refsinfo[r], word)
		}
	}

	fmt.Println(len(refsinfo))
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
	sort.Slice(refs, func(i, j int) bool { return refs[i].c > refs[j].c })
	for _, ref := range refs {
		words := refsinfo[ref.w]
		if len(words) == 1 {
			continue
		}
		var trans string
		if d := indexes.dictionary[ref.w]; d != nil {
			trans = d.Translation
		}
		fmt.Fprintf(out, "### %s(%d)\n", ref.w, ref.c)
		if trans == "" {
			var etym = indexes.etymology[ref.w]
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
}

func TestGenerateCoca2000(t *testing.T) {
	content, err := ioutil.ReadFile("C:/devs/go/src/etym/assets/coca20000.txt")
	if err != nil {
		panic(err)
	}
	words := strings.Split(string(content), "\n")
	for i := range words {
		words[i] = strings.TrimSpace(words[i])
	}
	//generateJSON(words, "C:/devs/go/src/etym/gen/coco20000.json")
	generateMD(words, "C:/devs/go/src/etym/gen/coco20000.md")
}
