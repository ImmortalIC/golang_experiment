package BestCacheInWorld

import (
	"sync"
	"time"
)

const (
	lifetime   = 5
	updatetime = 2
)

//ICache interface for cache module
type ICache interface {
	Get(key string, getter func() (interface{}, error)) (interface{}, error)
}

//Cache struct with data of cahch
type cache struct {
	maplock     sync.RWMutex
	dataStorage map[string]*cacheStorage
}

//Createcache factory for cache object
func Createcache() *cache {
	obj := new(cache)
	obj.dataStorage = make(map[string]*cacheStorage)
	return obj
}

type cacheStorage struct {
	lock       sync.RWMutex
	deathtime  int64
	data       interface{}
	cacheError error
}

func (c *cache) Get(key string, getter func() (interface{}, error)) (interface{}, error) {

	c.maplock.RLock()
	storage, storageExists := c.dataStorage[key]
	c.maplock.RUnlock()
	if !storageExists {
		storage = c.createKey(key)

	}
	storage.lock.RLock()

	if storage.deathtime > time.Now().Unix() {
		defer storage.lock.RUnlock()
		return storage.data, storage.cacheError
	}
	storage.lock.RUnlock()
	return storage.writeCache(getter)

}

func (c *cache) createKey(key string) *cacheStorage {
	c.maplock.Lock()
	defer c.maplock.Unlock()
	storage, storageExists := c.dataStorage[key]
	if !storageExists {
		c.dataStorage[key] = &cacheStorage{}
		storage = c.dataStorage[key]
	}
	return storage
}

func (storage *cacheStorage) writeCache(getter func() (interface{}, error)) (interface{}, error) {
	defer storage.lock.Unlock()
	storage.lock.Lock()
	if storage.deathtime <= time.Now().Unix() {
		storage.data, storage.cacheError = getter()

		storage.deathtime = time.Now().Unix() + lifetime
		go storage.updateCache(getter)
	}
	return storage.data, storage.cacheError
}

func (storage *cacheStorage) updateCache(getter func() (interface{}, error)) {
	time.Sleep(updatetime * time.Second)
	storage.lock.Lock()
	defer storage.lock.Unlock()
	storage.data, storage.cacheError = getter()
	if storage.deathtime > time.Now().Unix()+updatetime {
		go storage.updateCache(getter)
	}
}
