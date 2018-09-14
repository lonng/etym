package main

import (
	"etym/cmd/spider/command"
	"etym/cmd/spider/config"
	"etym/pkg/constant"
	"etym/pkg/fs"
	"etym/pkg/log"
	"flag"
	"path/filepath"
)

func main() {
	flag.StringVar(&config.SavePath, "save", "statics", "web content save path")
	flag.StringVar(&config.Word, "word", "", "fetch single word")
	flag.BoolVar(&config.Export, "export", false, "export json")
	flag.BoolVar(&config.EtymImg, "img", false, "export etymology image")
	flag.BoolVar(&config.Translate, "translate", false, "export translate")
	flag.BoolVar(&config.Reverse, "r", false, "reverse")
	flag.BoolVar(&config.EnableProxy, "proxy", false, "use proxy")
	flag.StringVar(&config.ProxyUrl, "url", "", "proxy url(http://host:port)")
	flag.IntVar(&config.Conc, "c", 5, "concurrency")
	flag.Parse()

	if config.SavePath == "" {
		log.Fatal("save path can not be empty")
	}

	log.SetLevel("debug")

	savePath, err := filepath.Abs(config.SavePath)
	if err != nil {
		log.Fatal(err)
	}
	savePath = filepath.Join(savePath, "etymology")
	config.SavePath = savePath
	log.Infof("Save path: %s", savePath)
	log.Infof("Concurrency: %d", config.Conc)

	fs.EnsureDir(savePath)
	fs.EnsureDir(filepath.Join(savePath, "indexes"))
	fs.EnsureDir(filepath.Join(savePath, "words"))
	fs.EnsureDir(filepath.Join(savePath, "export"))
	fs.EnsureDir(filepath.Join(savePath, constant.TranslationVer))

	if config.Export {
		// export to json
		command.ExportJson()
	} else if config.EtymImg {
		command.ExportImg()
	} else if config.Translate {
		command.ExportTrans()
	} else if word := config.Word; word != "" {
		// download single word
		command.DownloadWord(word)
	} else {
		// download indexes
		command.DownloadIndexes()
		command.DownloadWords()
	}

}
