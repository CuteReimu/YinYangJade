package db

import (
	"log/slog"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/pkg/errors"
)

// Set 设置键值对，ttl是超时时间（可选）
func Set(key, value string, ttl ...time.Duration) {
	err := DB.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), []byte(value))
		if len(ttl) > 0 {
			e = e.WithTTL(ttl[0])
		}
		err := txn.SetEntry(e)
		return err
	})
	if err != nil {
		slog.Error("set key value failed", "error", err, "key", key, "value", value)
	}
}

// Del 删除Key
func Del(key string) {
	err := DB.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		return err
	})
	if err != nil {
		slog.Error("delete key failed", "error", err, "key", key)
	}
}

// Get 根据Key获取Value
func Get(key string) (value string, ok bool) {
	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		} else if err != nil {
			return errors.Wrapf(err, "get failed, key: %s", key)
		}
		buf, err := item.ValueCopy(nil)
		if err == nil {
			value, ok = string(buf), true
		}
		return err
	})
	if err != nil {
		slog.Error("get failed", "error", err, "key", key)
	}
	return value, ok
}
