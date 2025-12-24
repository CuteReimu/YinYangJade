// Package db 用于初始化和管理 Badger 数据库连接。
package db

import (
	"log/slog"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var DB *badger.DB

func Init() {
	var err error
	DB, err = badger.Open(badger.DefaultOptions("assets/database"))
	if err != nil {
		panic("init database failed")
	}
	go gc()
}

func gc() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
	again:
		err := DB.RunValueLogGC(0.5)
		if err == nil {
			goto again
		}
	}
}

func Stop() {
	err := DB.Close()
	if err != nil {
		slog.Error("close database failed", "error", err)
	}
}
