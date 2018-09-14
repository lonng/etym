package main

import (
	"context"
	"etym/cmd/etymd/middleware"
	"etym/cmd/etymd/reviewd/api"
	"etym/pkg/db"
	"etym/pkg/errutil"
	"etym/pkg/fs"
	"etym/pkg/gziphandler"
	"etym/pkg/log"
	"etym/pkg/token"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lonnng/nex"
	"github.com/spf13/viper"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"
)

var staticDir string

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

	api.SetPrivateKeyPath(viper.GetString("login.private_key"))

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

	var (
		addr       = viper.GetString("webserver.review_addr")
		_staticDir = viper.GetString("webserver.static_dir")
	)

	_staticDir, err = filepath.Abs(_staticDir)
	if err != nil {
		panic(err)
	}

	if !fs.IsExists(_staticDir) {
		panic(fmt.Errorf("statics dir not found: %s", _staticDir))
	}

	staticDir = _staticDir

	log.Infof("Review service address: %s", addr)
	log.Infof("Statics directory: %v", _staticDir)
	router := mux.NewRouter()

	nex.Before(middleware.Logrequest)

	// Static files
	router.PathPrefix("/statics/").Handler(http.StripPrefix("/statics/", http.FileServer(http.Dir(_staticDir))))

	// Pages
	router.HandleFunc("/login", login)
	router.HandleFunc("/review", review)
	router.HandleFunc("/edit", edit)

	// API
	router.Handle("/api/v1/login", nex.Handler(api.LoginHandler)).Methods(http.MethodPost)
	router.Handle("/api/v1/improve/review", nex.Handler(api.RandReview).Before(checkAuth)).Methods(http.MethodGet)
	router.Handle("/api/v1/improve/reject", nex.Handler(api.RejectImprove).Before(checkAuth)).Methods(http.MethodPost)
	router.Handle("/api/v1/improve/edit", nex.Handler(api.RandEdit).Before(checkAuth)).Methods(http.MethodGet)
	router.Handle("/api/v1/improve/edit", nex.Handler(api.SubmitEdit).Before(checkAuth)).Methods(http.MethodPost)

	// PROXY
	router.HandleFunc("/api/v1/review/approve", api.Proxy)

	// Private API
	router.Handle("/api/v1/user/add", nex.Handler(api.AddUser).Before(middleware.LocalFilter)).Methods(http.MethodGet)

	api.RemoteUrl = fmt.Sprintf("http://%s", viper.GetString("webserver.rpc_addr"))
	panic(http.ListenAndServe(addr, gziphandler.GzipHandler(router)))
}

func checkAuth(ctx context.Context, request *http.Request) (context.Context, error) {
	if err := request.ParseForm(); err != nil {
		return ctx, err
	}

	t := request.Form.Get("token")
	if t == "" {
		return ctx, errutil.ErrPermissionDenied
	}

	meta, err := token.Parse(t)
	if err != nil {
		return ctx, errutil.ErrPermissionDenied
	}

	return context.WithValue(ctx, "account", meta.Account), nil
}

func login(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(staticDir, "views/login.html"))
}
func review(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(staticDir, "views/review.html"))
}
func edit(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(staticDir, "views/edit.html"))
}
