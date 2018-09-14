package command

import (
	"errors"
	"etym/cmd/spider/config"
	"etym/pkg/fs"
	"etym/pkg/log"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func save(rawurl, indexFile string) error {
	resp, err := http.Get(rawurl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	file, err := os.OpenFile(indexFile, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func indexFile(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}
	path := strings.Replace(strings.Trim(u.Path, "/"), "/", "_", -1)
	return filepath.Join(config.SavePath, "indexes", filepath.Clean(path+"#"+u.RawQuery+".html"))
}

func indexUrl(page int, q string) string {
	return fmt.Sprintf("https://www.etymonline.com/search?page=%d&q=%s", page, q)
}

func downloadIndex(page int, q string) {
	rawurl := indexUrl(page, q)
	indexFile := indexFile(rawurl)
	if fs.IsExists(indexFile) {
		return
	}
	log.Infof("Downloading URL=%s, savapath=%s", rawurl, indexFile)
	if err := save(rawurl, indexFile); err != nil {
		log.Error(err)
	}
}

func DownloadIndexes() {
	log.Infof("Starting download indexes...")
	for i := 'a'; i <= 'z'; i++ {
		downloadIndex(1, string(i))

		rawurl := indexUrl(1, string(i))
		indexFile := indexFile(rawurl)
		if !fs.IsExists(indexFile) {
			continue
		}

		n, err := matchCount(indexFile)
		if err != nil {
			log.Error(err)
			continue
		}

		page := int(math.Ceil(float64(n) / 10))
		log.Debugf("Entities Alphabetical=%s, count=>%d, page count=> %d", string(i), n, page)
		for j := 2; j <= page; j++ {
			downloadIndex(j, string(i))
		}
	}
	log.Infof("Download indexes completed.")
}

func matchCount(savefile string) (int, error) {
	data, err := ioutil.ReadFile(savefile)
	if err != nil {
		return 0, err
	}
	match := entitiesCounter.FindSubmatch(data)
	if len(match) < 2 {
		return 0, fmt.Errorf("match nothing")
	}
	return strconv.Atoi(string(match[1]))
}

func wordUrl(word string) string {
	return fmt.Sprintf("https://www.etymonline.com/word/%s", word)
}

func wordFile(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}
	path := strings.Replace(strings.Trim(u.Path, "/"), "/", "_", -1)
	return filepath.Join(config.SavePath, "words", filepath.Clean(path+".html"))
}

func DownloadWord(word string) {
	rawurl := wordUrl(word)
	indexFile := wordFile(rawurl)
	if fs.IsExists(indexFile) {
		return
	}
	//log.Infof("Downloading word URL=%s, savapath=%s", rawurl, indexFile)
	if err := save(rawurl, indexFile); err != nil {
		log.Warnf("Downloading word: %s, Error: %v", word, err)
	}
}

func DownloadWords() {
	log.Infof("Starting download words...")

	// 并发下载
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
	filepath.Walk(filepath.Join(config.SavePath, "indexes"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		result := word.FindAllSubmatch(data, -1)
		if len(result) < 1 {
			return nil
		}

		//log.Debugf("Find words => %s", path)
		for _, r := range result {
			word := string(r[1])
			if index := strings.Index(word, "("); index > 0 {
				word = word[:index]
			}
			word = strings.TrimSpace(word)
			word = html.UnescapeString(word)
			word = strings.Replace(word, "/", "%2F", -1)
			wg.Add(1)
			count++
			if count%100 == 0 {
				log.Infof("Downloading words count: %d", count)
			}
			tasks <- func() { DownloadWord(word) }
		}
		return nil
	})
	wg.Wait()
	close(tasks)
	log.Infof("Download word count: %d", count)
	log.Infof("Download words completed.")
}
