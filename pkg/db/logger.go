package db

import (
	"etym/pkg/log"
	"github.com/go-xorm/core"
)

type Logger struct {
	*log.Logger
	level core.LogLevel
}

func (l *Logger) SetLevel(level core.LogLevel) {
	l.level = level
}

func (l *Logger) Level() core.LogLevel {
	return l.level
}

func (l *Logger) ShowSQL(show ...bool) {}
func (l *Logger) IsShowSQL() bool      { return false }
