package db

import (
	"github.com/CuteReimu/dets"
	"github.com/dgraph-io/badger/v4"
	"log/slog"
	"time"
)

var DB *badger.DB

func Init() {
	var err error
	DB, err = badger.Open(badger.DefaultOptions("assets/database"))
	if err != nil {
		slog.Error("init database failed", "error", err)
		panic(err)
	}
	dets.SetDB(DB, slog.Default())
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
