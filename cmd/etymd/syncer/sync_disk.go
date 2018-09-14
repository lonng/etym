package syncer

import (
	"encoding/json"
	"etym/pkg/db/model"
	"etym/pkg/log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Record struct {
	Word  string
	Trans string
	Final bool
}

var (
	loadOnce sync.Once

	startFlag  int32
	stopFlag   int32
	dieChan    chan struct{}
	finishChan chan struct{}
	syncChan   chan Record
	dirtyCount int

	// shared data
	translationDict  = map[string]*model.PureTranslate{}
	translationSlice []*model.PureTranslate
	savePath         string
)

func init() {
	dieChan = make(chan struct{})
	finishChan = make(chan struct{})
}

func save() {
	if dirtyCount < 1 {
		return
	}

	bak := filepath.Join(filepath.Dir(savePath), "trans_bak.json")
	if err := os.Rename(savePath, bak); err != nil {
		log.Errorf("Rename: %v", err)
		return
	}

	saveFile, err := os.OpenFile(savePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Errorf("Create tmp translation file failed: %v", err)
		return
	}
	defer saveFile.Close()

	if err := json.NewEncoder(saveFile).Encode(translationSlice); err != nil {
		log.Errorf("Update translation file failed: %v", err)
	}

	log.Infof("Update translation files completed, Count=%d", dirtyCount)
	dirtyCount = 0
}

func Load(dataDir string, transFile string) {
	var load = func() {
		var transPath = filepath.Join(dataDir, transFile)
		log.Infof("Translation file path: %s", transPath)

		transFile, err := os.OpenFile(transPath, os.O_RDONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer transFile.Close()

		// 加载翻译
		log.Infof("Starting load translation: %s", transPath)
		if err := json.NewDecoder(transFile).Decode(&translationSlice); err != nil {
			panic(err)
		}
		for _, t := range translationSlice {
			translationDict[t.Word] = t
		}
		log.Infof("Loading translation complete")

		savePath = transPath
	}

	loadOnce.Do(load)
}

func Start() {
	if atomic.AddInt32(&startFlag, 1) != 1 {
		return
	}

	syncChan = make(chan Record, 1<<6)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)

		defer func() {
			ticker.Stop()
			close(syncChan)
			close(finishChan)
		}()

		for {
			select {
			case <-dieChan:
				save()
				log.Infof("Syncer stopping, save translation to trans.json")
				return

			case record := <-syncChan:
				etym, found := translationDict[record.Word]
				if !found || etym == nil {
					continue
				}

				if etym.Trans == record.Trans {
					continue
				}

				dirtyCount++
				etym.Trans = record.Trans
				etym.Final = record.Final

			case <-ticker.C:
				save()
			}
		}
	}()
}

func Sync(record Record) {
	if atomic.LoadInt32(&stopFlag) > 0 {
		return
	}
	syncChan <- record
}

func Find(word string) (string, bool) {
	trans, found := translationDict[word]
	if !found || trans == nil {
		return "", false
	}
	return trans.Trans, trans.Final
}

func Stop() {
	if atomic.LoadInt32(&startFlag) < 0 {
		log.Error("Syncer was not running")
		return
	}

	if atomic.AddInt32(&stopFlag, 1) != 1 {
		return
	}

	close(dieChan)
	<-finishChan
}
