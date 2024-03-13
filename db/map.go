package db

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/pkg/errors"
	"log/slog"
	"time"
)

// UpdateWithTtl 同 Update ，但是有一个ttl
func UpdateWithTtl(key []byte, f func(oldValue []byte) ([]byte, time.Duration)) (exists bool) {
	err := DB.Update(func(txn *badger.Txn) error {
		var newValue []byte
		var ttl time.Duration
		item, err := txn.Get(key)
		if errors.Is(err, badger.ErrKeyNotFound) {
			newValue, ttl = f(nil)
		} else if err != nil {
			return errors.Wrapf(err, "update key failed, key: %s", string(key))
		} else {
			exists = true
			err = item.Value(func(val []byte) error {
				newValue, ttl = f(val)
				return nil
			})
			if err != nil {
				return errors.Wrapf(err, "update key failed, key: %s", string(key))
			}
		}
		if newValue == nil {
			return nil
		} else if len(newValue) == 0 {
			err = txn.Delete(key)
		} else {
			e := badger.NewEntry(key, newValue).WithTTL(ttl)
			err = txn.SetEntry(e)
		}
		return err
	})
	if err != nil {
		slog.Error("update key failed, key: "+string(key), "error", err)
	}
	return
}

// Update 查找并修改值，返回原先是否存在。若原先不存在，则f的参数为nil。若f的返回值为nil，表示不进行Update。若f的返回值为空byte数组，表示删除。
func Update(key []byte, f func(oldValue []byte) []byte) (exists bool) {
	err := DB.Update(func(txn *badger.Txn) error {
		var newValue []byte
		item, err := txn.Get(key)
		if errors.Is(err, badger.ErrKeyNotFound) {
			newValue = f(nil)
		} else if err != nil {
			return errors.Wrapf(err, "update key failed, key: %s", string(key))
		} else {
			exists = true
			err = item.Value(func(val []byte) error {
				newValue = f(val)
				return nil
			})
			if err != nil {
				return errors.Wrapf(err, "update key failed, key: %s", string(key))
			}
		}
		if newValue == nil {
			return nil
		} else if len(newValue) == 0 {
			err = txn.Delete(key)
		} else {
			err = txn.Set(key, newValue)
		}
		return err
	})
	if err != nil {
		slog.Error("update key failed, key: "+string(key), "error", err)
	}
	return
}

// Set 设置键值对，ttl是超时时间（可选）
func Set(key, value []byte, ttl ...time.Duration) {
	err := DB.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, value)
		if len(ttl) > 0 {
			e = e.WithTTL(ttl[0])
		}
		err := txn.SetEntry(e)
		return err
	})
	if err != nil {
		slog.Error(fmt.Sprintf("set key value failed: {%s, %s}", string(key), string(value)), "error", err)
	}
}

// Del 删除Key
func Del(key []byte) {
	err := DB.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		return err
	})
	if err != nil {
		slog.Error("delete key failed, key: "+string(key), "error", err)
	}
}

// Get 根据Key获取Value
func Get(key []byte) (value []byte) {
	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		} else if err != nil {
			return errors.Wrapf(err, "get failed, key: %s", string(key))
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		slog.Error("get failed, key: "+string(key), "error", err)
	}
	return
}

func PrefixScanKeyValue(prefix []byte, f func(key, value []byte) error) {
	err := DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{PrefetchSize: 100})
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = f(k, v)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		slog.Error("prefix scan key failed: "+string(prefix), "error", err)
	}
}

func PrefixUpdateKey(prefix []byte, f func(key []byte) ([]byte, error)) {
	err := DB.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := f(k)
			if err != nil {
				return err
			}
			err = txn.Set(k, v)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		slog.Error("prefix update key failed: "+string(prefix), "error", err)
	}
}
