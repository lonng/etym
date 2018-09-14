package db

import (
	"etym/pkg/algoutil"
	"etym/pkg/constant"
	"etym/pkg/db/model"
	"etym/pkg/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"sync"
	"time"
)

const (
	defaultMaxConns = 10
)

type (
	options struct {
		showSQL      bool
		maxOpenConns int
		maxIdleConns int
	}

	Option func(opt *options)
	Closer func()
)

var (
	database *xorm.Engine // database connection
	initOnce sync.Once
)

// MaxIdleConns specifies the max idle connect numbers.
func MaxIdleConns(i int) Option {
	return func(opts *options) {
		opts.maxIdleConns = i
	}
}

// MaxOpenConns specifies the max open connect numbers.
func MaxOpenConns(i int) Option {
	return func(opts *options) {
		opts.maxOpenConns = i
	}
}

// ShowSQL specifies the buffer size.
func ShowSQL(show bool) Option {
	return func(opts *options) {
		opts.showSQL = show
	}
}

func Initialize(dsn string, opts ...Option) {
	var connect = func() {
		var settings = &options{
			maxIdleConns: defaultMaxConns,
			maxOpenConns: defaultMaxConns,
			showSQL:      false,
		}
		for _, o := range opts {
			o(settings)
		}

		log.Infof("DB source: %s", dsn)
		// create database instance
		db, err := xorm.NewEngine("mysql", dsn)
		if err != nil {
			panic(err)
		}
		database = db

		// options
		database.SetMaxIdleConns(settings.maxIdleConns)
		database.SetMaxOpenConns(settings.maxOpenConns)
		database.ShowSQL(settings.showSQL)
		database.SetLogger(&Logger{Logger: log.Std()})

		if err := database.Ping(); err != nil {
			panic(err)
		}

		syncSchema()
		startPing()
		checkSuperadmin()
	}
	initOnce.Do(connect)
}

func Database() *xorm.Engine {
	return database
}

func startPing() {
	// 定时ping数据库, 保持连接池连接
	go func() {
		ticker := time.NewTicker(time.Minute * 5)
		for {
			select {
			case <-ticker.C:
				if err := database.Ping(); err != nil {
					log.Error(err)
				}
			}
		}
	}()
}

func checkSuperadmin() {
	if !IsAccountExist(constant.Superadmin) {
		log.Info("superadmin account not found in database, will be create automatically")
		const username = constant.Superadmin
		const password = "xxsuperadminxxl"
		hash, salt := algoutil.PasswordHash(password)
		u := &model.DbUser{
			Name:          username,
			Account:       username,
			Password:      hash,
			Salt:          salt,
			Role:          model.RoleSuperAdmin,
			Status:        model.AgentStatusActivated,
			CreateAccount: username,
			Extra:         "CREATE AUTOMATICALLY",
		}

		if err := RegisterUser(u); err != nil {
			log.Error(err)
		}
	}
}
