package utils

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/storage/redis"
)

func SetCache(storage *redis.Storage, key string, data interface{}, duration time.Duration) error {
	if storage == nil {
		return errors.New("redis storage is nil")
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return storage.Set(key, jsonData, duration)
}
func GetCache(storage *redis.Storage, key string) (interface{}, error) {
	if storage == nil {
		return nil, errors.New("redis storage is nil")
	}
	data, err := storage.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errors.New("no data found in cache")
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func DeleteCache(storage *redis.Storage, key string) error {
	if storage == nil {
		return errors.New("redis storage is nil")
	}
	return storage.Delete(key)
}
func GetCacheOrSetIfNotExist(storage *redis.Storage, key string, duration time.Duration, fetchFunc func() (interface{}, error)) (interface{}, error) {
	if storage == nil {
		return nil, errors.New("redis storage is nil")
	}
	cachedData, err := GetCache(storage, key)
	if err == nil && cachedData != nil {
		return cachedData, nil
	}
	data, err := fetchFunc()
	if err != nil {
		return nil, err
	}
	err = SetCache(storage, key, data, duration)
	if err != nil {
		return nil, err
	}
	return data, nil
}
