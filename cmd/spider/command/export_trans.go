package command

import (
	"encoding/json"
	"errors"
	"etym/cmd/spider/config"
	"etym/cmd/spider/model"
	"etym/pkg/constant"
	etymodel "etym/pkg/db/model"
	"etym/pkg/fs"
	"etym/pkg/log"
	"etym/pkg/proxy"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

func ExportTrans() {
	log.Infof("Starting export translate...")

	// 代理列表锁
	var addrChan chan string
	if config.EnableProxy {
		addrChan = make(chan string, 32)
		go func() {
			log.Infof("Starting feching proxy list")
			for {
				err := proxy.Get(addrChan)
				if err != nil {
					log.Warnf("Fetching proxy failed: %s", err.Error())
					continue
				}
			}
		}()
	}

	var defaultClient = http.DefaultClient
	if config.ProxyUrl != "" {
		proxyUrl, err := url.Parse(config.ProxyUrl)
		if err == nil {
			defaultClient = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyUrl),
				},
				Timeout: 5 * time.Second,
			}
		}
	}
	var getHttpClient = func() *http.Client {
		if !config.EnableProxy {
			return defaultClient
		}

		addr := <-addrChan
		proxyUrl, err := url.Parse(addr)
		if err != nil {
			return defaultClient
		}

		log.Infof("Construct new proxy client, addr=%v", addr)
		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			},
			Timeout: proxy.Timeout * time.Second,
		}
	}

	var exportPath = func(word string) string {
		word = strings.Replace(strings.Trim(word, "/"), "/", "_", -1)
		return filepath.Join(config.SavePath, constant.TranslationVer, fmt.Sprintf("%s.json", strings.TrimSpace(word)))
	}

	var translate = func(client *http.Client, sentence string) (*etymodel.SpiderTranslate, error) {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			return nil, nil
		}
		query := url.Values{}
		query.Set("q", sentence)
		resp, err := client.PostForm(google, query)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(resp.Status)
		}

		var trans = &etymodel.SpiderTranslate{}
		if err := json.NewDecoder(resp.Body).Decode(trans); err != nil {
			return nil, fmt.Errorf("json decode: %v", err)
		}
		if len(trans.Sentences) < 1 {
			return nil, errors.New("empty translate sentences response")
		}
		return trans, nil
	}

	var export = func(client *http.Client, word, etym string) error {
		path := exportPath(word)
		if fs.IsExists(path) {
			return nil
		}

		log.Debugf("Starting translate, Word=%s", word)
		trans, err := translate(client, etym)
		if err != nil {
			return err
		}

		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("open save file %s failed: %s", path, err.Error())
		}
		defer file.Close()

		log.Debugf("Complete translate, Word=%s, Path=%s", word, path)
		return json.NewEncoder(file).Encode(&model.Word{
			Word:  word,
			Trans: trans,
		})
	}

	// 并发导出
	var wg sync.WaitGroup
	type Task func(*http.Client) error
	var tasks = make(chan Task, 1<<5)
	for i := 0; i < config.Conc; i++ {
		go func(no int) {
			client := getHttpClient()
			log.Infof("Gorontine %d get client success, start working.", no)
			var failCount int
			for t := range tasks {
				if err := t(client); err != nil {
					failCount++
					// switch client
					if failCount >= 30 {
						failCount = 0
						client = getHttpClient()
						log.Infof("Gorontine %d switch client success.", no)
					}

					log.Warnf("Execute task failure: %v", err)
					// pending retry
					wg.Add(1)
					tasks <- t
				}
				wg.Done()
			}
		}(i)
	}

	var count int
	var fail int
	var files []string
	filepath.Walk(filepath.Join(config.SavePath, "export"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})

	if config.Reverse {
		log.Infof("Reverse fetching translate.")
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
	} else {
		sort.Strings(files)
	}

	const trimLineBreak = true // 是否处理符号平衡问题

	// 替换c. 1800为 c.1800[,.;\s]
	centRE := regexp.MustCompile(`[c|C]\.\s+(\d{3,4})`)
	x2re := regexp.MustCompile(`\W[c|C]\.\s+(\d{3,4})`)
	_ = x2re
	var replcount int
	for _, path := range files {
		file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Warnf("Open file %v failed: %v", path, err)
			continue
		}
		etym := new(model.Etym)
		if err := json.NewDecoder(file).Decode(etym); err != nil {
			log.Warnf("Decode error: %v", err)
			continue
		}

		if trimLineBreak {
			lines := strings.Split(etym.Etym, "\n")
			for i := 0; i < len(lines); i++ {
				lines[i] = strings.TrimSpace(lines[i])
				if lines[i] == "" {
					if i == len(lines)-1 {
						lines = lines[:i]
					} else {
						lines = append(lines[:i], lines[i+1:]...)
					}
				}
			}
			etym.Etym = strings.Join(lines, " ")
		}

		/*x1 := centRE.Find([]byte(etym.Etym))
		x2 := x2re.Find([]byte(etym.Etym))
		if len(x1) > 0 && len(x2) < 1 {
			//log.Infof("+>%s", x1)
			//log.Infof("=>%s", etym.Etym)
			path := exportPath(etym.Word)
			if fs.IsExists(path) {
				log.Infof("删除=>Word=%v, ERROR:%v, PATH=%s", etym.Word, os.Remove(path), path)
			}
		}*/

		// 处理世纪
		x := centRE.Find([]byte(etym.Etym))
		if len(x) > 0 {
			replcount++
			repl := centRE.ReplaceAll([]byte(etym.Etym), []byte("around $1"))
			//log.Infof("+>%s", x)
			//log.Infof("=>%s", etym.Etym)
			//log.Infof("->%s", string(repl))
			etym.Etym = string(repl)
			//path := exportPath(etym.Word)
			//if fs.IsExists(path) {
			//	log.Infof("删除=>Word=%v, ERROR:%v, PATH=%s", etym.Word, os.Remove(path), path)
			//}
		}

		//log.Infof("Word=%v, Etym=%v", word, et]]]]]ym)
		wg.Add(1)
		count++
		if count%100 == 0 {
			log.Infof("Exporting translate count: %d", count)
		}
		task := func(closureWord, closureEtym string) Task {
			return func(client *http.Client) error {
				return export(client, closureWord, closureEtym)
			}
		}(etym.Word, etym.Etym)
		tasks <- task
	}

	wg.Wait()
	close(tasks)
	log.Infof("Exporting translate completed, count=%d, fail=%v, repl=%d.", count, fail, replcount)
}
