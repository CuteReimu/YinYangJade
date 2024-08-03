package db

import (
	"github.com/dgraph-io/badger/v4"
	"log/slog"
	"time"
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
