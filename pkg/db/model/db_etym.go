package model

import (
	"time"
)

type (
	DbUser struct {
		Id             int64  `xorm:"BIGINT(20) NOT NULL"`
		Name           string `xorm:"VARCHAR(32) NOT NULL"`
		Account        string `xorm:"VARCHAR(32) NOT NULL"`
		Password       string `xorm:"VARCHAR(64) NOT NULL"`
		Salt           string `xorm:"VARCHAR(32) NOT NULL"`
		Role           int    `xorm:"TINYINT(4) NOT NULL"`
		Status         int    `xorm:"TINYINT(4) NOT NULL"`
		Extra          string `xorm:"VARCHAR(255) NOT NULL"`
		CreateAt       int64  `xorm:"BIGINT(20) NOT NULL"`
		DeleteAt       int64  `xorm:"BIGINT(20) NOT NULL"`
		DeleteAccount  string `xorm:"VARCHAR(32) NOT NULL"`
		CreateAccount  string `xorm:"VARCHAR(32) NOT NULL"`
		ConfirmAccount string `xorm:"VARCHAR(32) NOT NULL"`
		Level          int    `xorm:"INT(20) NOT NULL"`
	}

	DbHistory struct {
		Id            int64
		Word          string    `xorm:"not null index VARCHAR(128)"` // 单词
		LowerWord     string    `xorm:"not null index VARCHAR(128)"` // 单词
		Original      string    `xorm:"not null TEXT"`               // 原文
		PrevHistoryId int64     `xorm:"not null bigint"`
		PrevTrans     string    `xorm:"not null TEXT"`       // 修改前
		NextTrans     string    `xorm:"not null TEXT"`       // 修改后
		ProposalCount int       `xorm:"not null TINYINT(3)"` // 修改数量
		Status        int       `xorm:"not null TINYINT(3)"` // 状态
		ReviewCount   int       `xorm:"not null default 0"`  // 审阅次数
		Score         int       `xorm:"not null default 0"`  // 评分
		CreateAccount string    `xorm:"not null VARCHAR(128)"`
		CreatedAt     time.Time `xorm:"not null created"`
		UpdatedAt     time.Time `xorm:"not null updated"`
	}
)
