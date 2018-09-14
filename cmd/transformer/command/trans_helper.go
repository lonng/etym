package command

import (
	"etym/pkg/db/model"
	"html"
	"regexp"
	"strings"
)

type CombineResult struct {
	Original    string
	Translation string
	Final       int
}

func combineSentences(sentences []*model.Sentence) CombineResult {
	var pivot = -1
	for i := range sentences {
		sentences[i].Trans = strings.TrimSpace(html.UnescapeString(sentences[i].Trans))
		sentences[i].Orig = strings.TrimSpace(html.UnescapeString(sentences[i].Orig))
		if strings.HasPrefix(sentences[i].Orig, "Related:") {
			pivot = i
		}
	}

	// 相关单词不做翻译
	if pivot >= 0 {
		var orig []string
		for j := pivot; j < len(sentences); j++ {
			orig = append(orig, sentences[j].Orig)
		}
		origstr := strings.Join(orig, " ")
		sentences = sentences[:pivot+1]
		sentences[pivot].Orig = origstr
		sentences[pivot].Trans = origstr
	}

	var origs []string
	var trans []string
	for _, s := range sentences {
		origs = append(origs, s.Orig)
		trans = append(trans, s.Trans)
	}

	combineOrigs := strings.Join(origs, " ")
	combineTrans := strings.Join(trans, " ")
	final := 0

	// see开头的不翻译
	if strings.HasPrefix(combineOrigs, "see") {
		combineTrans = strings.Replace(combineOrigs, "see", "参阅", 1)
		final = 1
	} else if strings.HasPrefix(combineOrigs, "See") {
		combineTrans = strings.Replace(combineOrigs, "See", "参阅", 1)
		final = 2
	} else {
		// 参考不翻译 (see xxxxx)
		seeOrigRE := regexp.MustCompile(`\(see\s.*?(\(\w+\.?.*?\))?.*?\)`)
		seeTranRE := regexp.MustCompile(`（.*?(（(\w+)。(.*?)）)?.*?）`)

		if xx := seeOrigRE.FindAll([]byte(combineOrigs), -1); len(xx) > 0 {
			indexes := seeTranRE.FindAllIndex([]byte(combineTrans), -1)
			if len(indexes) != len(xx) {
				// 保守策略, 这种情况直接使用原文
				//combineTrans = combineOrigs
			} else {
				lastIndex := indexes[len(indexes)-1][1]
				var trans string
				//log.Infof("+=> %s", sentences[i].Orig)
				//log.Infof("+=> %s", sentences[i].Trans)
				for j, index := range indexes {
					trans += combineTrans[:index[0]]
					trans += string(xx[j])
				}
				trans += combineTrans[lastIndex:]
				//log.Infof("+=> %s", trans)
				//sentences[i].Trans = string(tranRE.ReplaceAll([]byte(sentences[i].Trans), []byte("($1.$2)")))
				//log.Infof("+=> %s", sentences[i].Trans)
				combineTrans = trans
			}
		}

		// 词性不翻译, (n.)被Google翻译为(年.)
		origRE := regexp.MustCompile(`\(\w+\..*?\)`)
		tranRE := regexp.MustCompile(`（(\w+)。(.*?)）`)
		if x := origRE.FindString(combineOrigs); len(x) > 0 {
			combineTrans = string(tranRE.ReplaceAllString(combineTrans, "($1.$2)"))
		}
	}

	return CombineResult{Original: combineOrigs, Translation: combineTrans, Final: final}
}
