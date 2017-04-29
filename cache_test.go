package BestCacheInWorld

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func startCaching(cache ICache, assertVal int64, wait int, key string, getter func() (interface{}, error), result chan bool) {
	time.Sleep(time.Duration(wait) * time.Millisecond)
	val, err := cache.Get(key, getter)

	result <- int64(val.(int)) == assertVal && err == nil
}

func TestFunctionality(t *testing.T) {

	cache := Createcache()

	rakiChaneru := make(chan bool)
	go startCaching(cache, 2, 0, "key1", func() (interface{}, error) {
		return 2, nil
	}, rakiChaneru)
	go startCaching(cache, 2, 0, "key1", func() (interface{}, error) {
		return 2, nil
	}, rakiChaneru)
	go startCaching(cache, 4, 0, "key2", func() (interface{}, error) {
		return 4, nil
	}, rakiChaneru)
	go startCaching(cache, 545, 0, "key3", func() (interface{}, error) {
		return 545, nil
	}, rakiChaneru)

	go startCaching(cache, 545, 2000, "key3", func() (interface{}, error) {
		return 545, nil
	}, rakiChaneru)
	total := 0
	success := 0
	var result bool
	for total < 5 {
		result = <-rakiChaneru
		fmt.Println(result)
		total++
		if result {
			success++
		}
	}

	if !(success == 5) {

		t.FailNow()

	}

}

func TestGetterusing(t *testing.T) {
	cache := Createcache()
	rakiChaneru := make(chan bool)
	getterInvoked := 0
	var counterLock sync.RWMutex
	getter := func() (interface{}, error) {
		counterLock.Lock()
		getterInvoked++
		counterLock.Unlock()
		return 2, nil
	}
	go startCaching(cache, 2, 0, "key1", getter, rakiChaneru)
	go startCaching(cache, 2, 0, "key1", getter, rakiChaneru)
	go startCaching(cache, 2, 0, "key1", getter, rakiChaneru)
	go startCaching(cache, 2, 5050, "key1", getter, rakiChaneru)
	total := 0
	success := 0
	var result bool
	for total < 3 {
		result = <-rakiChaneru
		fmt.Println(result)
		total++
		if result {
			success++
		}
	}
	defer counterLock.RUnlock()
	counterLock.RLock()
	if success != 3 || getterInvoked > 1 {
		fmt.Println(success, getterInvoked)
		t.FailNow()
	}
	counterLock.RUnlock()
	time.Sleep(2500 * time.Millisecond)
	counterLock.RLock()
	if getterInvoked != 2 {
		fmt.Println("no update happened or too many", getterInvoked)
		t.FailNow()
	}
	counterLock.RUnlock()
	result = <-rakiChaneru
	counterLock.RLock()
	if !result || getterInvoked != 4 {
		fmt.Println("no update or invalidation happened or too many", getterInvoked, result)
		t.FailNow()
	}

}

type err4Test struct {
}

func (e err4Test) Error() string {
	return "%%error%%"
}

func TestErrorreporting(t *testing.T) {
	cache := Createcache()
	rakiChaneru := make(chan bool)
	timestamp := time.Now().Unix()
	getter := func() (interface{}, error) {
		if (time.Now().Unix() - timestamp) < 2 {
			return 0, err4Test{}
		} else {
			return 5, nil
		}
	}
	errTester := func() {
		val, err := cache.Get("key1", getter)

		rakiChaneru <- err.Error() == "%%error%%" && val == 0
	}
	for i := 0; i < 4; i++ {
		go errTester()
	}
	for i := 0; i < 4; i++ {
		go startCaching(cache, 5, 2100, "key1", getter, rakiChaneru)
	}
	total := 0
	success := 0
	var result bool
	for total < 8 {
		result = <-rakiChaneru
		fmt.Println(result)
		total++
		if result {
			success++
		}
	}
	if success < 8 {
		fmt.Println("Not all threads goes right", success)
		t.FailNow()
	}

}
