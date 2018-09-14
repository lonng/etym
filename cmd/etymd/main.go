package main

import (
	"etym/pkg/db"
	"etym/pkg/fs"
	"etym/pkg/gziphandler"
	"etym/pkg/log"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lonnng/nex"
	"github.com/spf13/viper"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"

	"etym/cmd/etymd/middleware"
	"etym/cmd/etymd/services/etym"
	"etym/cmd/etymd/services/etym/memdb"
	"etym/cmd/etymd/services/etym/wordlist"
	"etym/cmd/etymd/services/review"
	"etym/cmd/etymd/syncer"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var configPath = "configs/etymd.config.toml"
	flag.StringVar(&configPath, "f", configPath, "specify config file path")
	flag.Parse()

	var err error
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		panic(err)
	}

	if !fs.IsExists(configPath) {
		panic(fmt.Errorf("config path not exists: %s", configPath))
	}

	// Loading config
	log.Infof("Config path: %s", configPath)
	viper.SetConfigType("toml")
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	var (
		enableSSL = viper.GetBool("webserver.enable_ssl")
		addr      = viper.GetString("webserver.addr")
		certFile  string
		keyFile   string
		dataDir   = viper.GetString("core.data_dir")
		rpcAddr   = viper.GetString("webserver.rpc_addr")
	)

	dataDir, err = filepath.Abs(dataDir)
	if err != nil {
		panic(err)
	}

	if !fs.IsExists(dataDir) {
		panic(fmt.Errorf("data directory not found: %s", dataDir))
	}

	options := &memdb.Options{
		EtymFile:  viper.GetString("core.data_files.etym"),
		DictFile:  viper.GetString("core.data_files.dict"),
		LemmaFile: viper.GetString("core.data_files.lemma"),
	}
	// strict order
	syncer.Load(dataDir, viper.GetString("core.data_files.trans")) // initialize syncer
	memdb.Load(dataDir, options)                                   // initialize memory database
	wordlist.Load(dataDir)                                         // initialize wordlist

	log.Infof("Enable SSL: %v", enableSSL)
	log.Infof("Listen address: %v, rpc: %s", addr, rpcAddr)
	if enableSSL {
		certFile = viper.GetString("webserver.certificates.cert")
		keyFile = viper.GetString("webserver.certificates.key")
		log.Infof("Cert file path: %s", certFile)
		log.Infof("Key file path: %s", keyFile)
	}

	// connect database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetString("database.port"),
		viper.GetString("database.dbname"),
		viper.GetString("database.args"))

	opts := []db.Option{
		db.MaxIdleConns(viper.GetInt("database.max_idle_conns")),
		db.MaxOpenConns(viper.GetInt("database.max_open_conns")),
		db.ShowSQL(viper.GetBool("database.show_sql")),
	}
	db.Initialize(dsn, opts...)

	nex.Before(middleware.Logrequest)

	// startup http service
	router := mux.NewRouter()

	// services
	router.Handle("/api/v1/review/improve", nex.Handler(review.ImproveInfo)).Methods(http.MethodGet)
	router.Handle("/api/v1/review/improve", nex.Handler(review.SubmitImprove)).Methods(http.MethodPost)

	// statics files
	//router.PathPrefix("/statics/").Handler(http.StripPrefix("/statics/", http.FileServer(http.Dir(staticDir))))
	router.PathPrefix("/.well-known/acme-challenge/").Handler(http.StripPrefix("/.well-known/acme-challenge/", http.FileServer(http.Dir("challenges"))))

	// API: memdb
	router.Handle("/api/v1/wordlist", nex.Handler(wordlist.WordList)).Methods(http.MethodGet) // 支持的wordlist
	router.Handle("/api/v1/rand", nex.Handler(wordlist.Rand)).Methods(http.MethodGet)         // 随机获取一个单词
	router.Handle("/api/v1/etym", nex.Handler(etym.Etym)).Methods(http.MethodGet)             // 词源信息
	router.Handle("/api/v1/search", nex.Handler(etym.Search)).Methods(http.MethodGet)         // 搜索词源

	// Public API server
	go func() {
		var err error
		server := &http.Server{Addr: addr, Handler: gziphandler.GzipHandler(router)}
		if enableSSL {
			err = server.ListenAndServeTLS(certFile, keyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil {
			panic(err)
		}
	}()

	// Private RPC API Server
	adminRouter := mux.NewRouter()
	adminRouter.Handle("/api/v1/review/approve", nex.Handler(review.Approve).Before(middleware.LocalFilter)).Methods(http.MethodGet)
	go func() {
		panic(http.ListenAndServe(rpcAddr, adminRouter))
	}()

	syncer.Start()

	// Shutdown server gracefully
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	s := <-signals
	log.Infof("Got signal: %s, system will shutdown soon", s.String())

	// Shutdown hooks
	syncer.Stop()

	log.Info("System shutdown complete, will exit soon")
}
