package command

import (
	"encoding/json"
	"etym/cmd/spider/config"
	"etym/cmd/spider/model"
	etymodel "etym/pkg/db/model"
	"etym/pkg/fs"
	"etym/pkg/log"
	"etym/pkg/template"
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

func ExportJson() {
	log.Infof("Starting export words...")
	var exportPath = func(word string) string {
		word = strings.Replace(strings.Trim(word, "/"), "/", "_", -1)
		return filepath.Join(config.SavePath, "export", fmt.Sprintf("%s.json", strings.TrimSpace(word)))
	}

	var crossref struct {
		sync.Mutex
		exported map[string]struct{}
		ref      map[string]struct{}
	}
	crossref.ref = map[string]struct{}{}
	crossref.exported = map[string]struct{}{}

	var export = func(word, etym string) {
		path := exportPath(word)
		if fs.IsExists(path) {
			return
		}
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Errorf("Open save file %s failed, Error=%s", path, err.Error())
			return
		}
		defer file.Close()

		fetym := html.UnescapeString(template.StripTags(etym))
		var refs, foreigns []string
		if result := ref.FindAllStringSubmatch(etym, -1); len(result) > 0 {
			for _, r := range result {
				refs = append(refs, r[1])
				crossref.Lock()
				crossref.ref[etymodel.CleanWord(r[1])] = struct{}{}
				crossref.Unlock()
			}
		}

		if result := foreign.FindAllStringSubmatch(etym, -1); len(result) > 0 {
			for _, r := range result {
				foreigns = append(foreigns, r[1])
				crossref.Lock()
				crossref.ref[etymodel.CleanWord(r[1])] = struct{}{}
				crossref.Unlock()
			}
		}

		//log.Debugf("导出单词, Word=%s, Ref=%+v, Foreign=%+v", word, refs, foreigns)
		err = json.NewEncoder(file).Encode(&model.Etym{
			Word:    word,
			RawEtym: etym,
			Etym:    fetym,
			Ref:     refs,
			Foreign: foreigns,
		})
		if err != nil {
			log.Errorf("Save file %s failed: %s", path, err.Error())
			return
		}
		crossref.Lock()
		crossref.exported[etymodel.CleanWord(word)] = struct{}{}
		crossref.Unlock()
	}

	// 并发导出
	var wg sync.WaitGroup
	var tasks = make(chan func(), 1<<5)
	for i := 0; i < config.Conc; i++ {
		go func() {
			for t := range tasks {
				t()
				wg.Done()
			}
		}()
	}

	var count int
	var fail int
	var files []string
	filepath.Walk(filepath.Join(config.SavePath, "words"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})

	sort.Strings(files)

	for _, path := range files {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Warnf("Open save file %s failed, Error=%s", path, err.Error())
			continue
		}
		result := exporter.FindAllSubmatch(data, -1)
		if len(result) < 1 {
			fail++
			log.Warnf("Match words failure: %v", path)
			continue
		}

		for _, r := range result {
			_ = r
			word := string(r[1])
			etym := string(r[2])

			//log.Infof("Word=%v, Etym=%v", word, etym)
			wg.Add(1)
			count++
			if count%1000 == 0 {
				log.Infof("Exporting words count: %d", count)
			}
			tasks <- func(closureWord, closureEtym string) func() { return func() { export(closureWord, closureEtym) } }(word, etym)
		}
	}

	wg.Wait()
	log.Infof("Exporting json completed, count=%d, fail=%v.", count, fail)

	const checkcross = false
	if checkcross {
		log.Infof("Check cross reference: %d", len(crossref.ref))
		var totalRef = len(crossref.ref)
		var currentRef int
		for word := range crossref.ref {
			currentRef++
			if _, found := crossref.exported[word]; found {
				continue
			}
			rawurl := wordUrl(word)
			indexFile := wordFile(rawurl)
			if fs.IsExists(indexFile) {
				continue
			}

			if currentRef%1000 == 0 {
				log.Infof("Download ref progress: %d/%d", currentRef, totalRef)
			}

			wg.Add(1)
			tasks <- func(w string) func() { return func() { DownloadWord(w) } }(word)
		}
		wg.Wait()
	}

	close(tasks)
}
