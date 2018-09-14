package command

import (
	"encoding/json"
	"errors"
	"etym/cmd/spider/config"
	"etym/cmd/spider/model"
	"etym/pkg/constant"
	"etym/pkg/fs"
	"etym/pkg/log"
	"etym/pkg/proxy"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func ExportImg() {
	log.Infof("Starting export original image...")

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
		return filepath.Join(config.SavePath, constant.ImageVer, fmt.Sprintf("%s.base64", strings.TrimSpace(word)))
	}

	const service = "https://www.google.com.hk/search?q=dictionary+"
	var fetchimg = func(client *http.Client, word string) ([]byte, error) {
		word = strings.TrimSpace(word)
		if word == "" {
			return nil, nil
		}

		resp, err := client.Get(service + word)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(resp.Status)
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		result := originImg.Find(data)

		fmt.Println(string(data))
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println(string(result))
		return result, nil
	}

	var export = func(client *http.Client, word string) error {
		path := exportPath(word)
		if fs.IsExists(path) {
			return nil
		}

		log.Debugf("Starting fetching origin img, Word=%s", word)
		base64img, err := fetchimg(client, word)
		if err != nil {
			return err
		}

		log.Debugf("Complete fetch image, Word=%s, Path=%s", word, path)
		return ioutil.WriteFile(path, base64img, os.ModePerm)
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
		log.Infof("Reverse fetching image.")
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
	} else {
		sort.Strings(files)
	}

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

		// 只获取纯字母的单词, 词根不获取

		wg.Add(1)
		count++
		if count%100 == 0 {
			log.Infof("Exporting translate count: %d", count)
		}
		task := func(closureWord string) Task {
			return func(client *http.Client) error {
				return export(client, closureWord)
			}
		}("hello")
		tasks <- task

		// TODO: test 1
		break
	}

	wg.Wait()
	close(tasks)
	log.Infof("Exporting image completed, count=%d, fail=%v", count, fail)
}
