package fodcache

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
)

/*******************************************************************************
 * This Source Code Form is copyright of 51Degrees Mobile Experts Limited.
 * Copyright 2017 51Degrees Mobile Experts Limited, 5 Charlotte Close,
 * Caversham, Reading, Berkshire, United Kingdom RG4 7BY
 *
 * This Source Code Form is the subject of the following patents and patent
 * applications, owned by 51Degrees Mobile Experts Limited of 5 Charlotte
 * Close, Caversham, Reading, Berkshire, United Kingdom RG4 7BY:
 * European Patent No. 2871816;
 * European Patent Application No. 17184134.9;
 * United States Patent Nos. 9,332,086 and 9,350,823; and
 * United States Patent Application No. 15/686,066.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0.
 *
 * If a copy of the MPL was not distributed with this file, You can obtain
 * one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, v. 2.0.
 ******************************************************************************/

func TestLruPutCache_Get(t *testing.T) {
	cb := CacheBuilder{CacheSize: 2}
	cache, err := NewLruPutCache(&cb)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = cache.Put(1, "test")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	result, err := cache.LruCacheBase.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	//strResult := fmt.Sprintf("%v",result)

	if result != "test" {
		t.Error(AssertError)
		t.FailNow()
	}
}

func TestLruPutCache_NoValue(t *testing.T) {
	cb := CacheBuilder{CacheSize: 2}
	cache, err := NewLruPutCache(&cb)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	result, err := cache.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	//strResult := fmt.Sprintf("%v",result)

	if result != nil {
		t.Error(AssertError)
		t.FailNow()
	}
}

func TestLruPutCache_LruPolicyCheck(t *testing.T) {

	cb := CacheBuilder{Concurrency: 1, CacheSize: 2}
	cache, err := NewLruPutCache(&cb)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Add three items in a row.
	_, err = cache.Put(1, "test1")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = cache.Put(2, "test2")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	_, err = cache.Put(3, "test3")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	result1, err := cache.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	result2, err := cache.Get(2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	result3, err := cache.Get(3)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// The oldest item should have been evicted.
	if result1 != nil {
		t.Log(fmt.Printf("result1 is not nil, it's %s", result1))
		t.Error(AssertError)
		t.FailNow()
	}
	if result2 != "test2" {
		t.Log(fmt.Printf("result2 is not test2, it's %s", result2))
		t.Error(AssertError)
		t.FailNow()
	}

	if result3 != "test3" {
		t.Log(fmt.Printf("result3 is not test3, it's %s", result3))
		t.Error(AssertError)
		t.FailNow()
	}
}

func TestLruPutCache_LruPolicyCheck2(t *testing.T) {
	cb := CacheBuilder{CacheSize: 2, Concurrency: 1}
	cache, err := NewLruPutCache(&cb)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Add two items.
	_, err = cache.Put(1, "test1")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = cache.Put(2, "test2")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Access the first one.
	_, err = cache.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Add a third item.
	_, err = cache.Put(3, "test3")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	result1, err := cache.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	result2, err := cache.Get(2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	result3, err := cache.Get(3)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// The second item should have been evicted.
	if result1 != "test1" {
		t.Log(fmt.Printf("result1 is not test1, it's %s", result1))
		t.Error("")
	}

	if result2 != nil {
		t.Log(fmt.Printf("result2 is not null, it's %s", result2))
		t.Error("")
	}

	if result3 != "test3" {
		t.Log(fmt.Printf("result3 is not test3, it's %s", result3))
		t.Error("")
	}
}

// ExecutionException, InterruptedException
func TestLruPutCache_HighConcurrency(t *testing.T) {

	var hits uint64 // this will be incremented atomically

	cb := CacheBuilder{CacheSize: 100, Concurrency: 50}
	cache, err := NewLruPutCache(&cb)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	wg := sync.WaitGroup{}
	totalRequests := 1000000
	var queue = make(chan int, totalRequests)
	//var start = make(chan bool, 1)
	rnd := Random{}
	for i := 0; i < totalRequests; i++ {
		queue <- rnd.nextInt(200)
	}

	for i := 0; i < 50; i++ {
		//for i := 0; i < 1; i++ {
		wg.Add(1)
		go func() {
			log.Println("Started")
			for {
				// this blocks until we fire the starting gun
				select {
				// non-blocking read of queue
				case key := <-queue:
					result, _ := cache.Get(key)
					if result == nil {
						// we're ignoring a returned CachedItem and a possible error here
						_, _ = cache.Put(key, "test"+string(key))
					} else {
						if fmt.Sprintf("%v", result) != ("test" + string(key)) {
							t.Error(AssertError)
							t.FailNow()
						}
						atomic.AddUint64(&hits, 1)
					}
				default:
					// we're done.
					wg.Done()
					return
				}
			}
		}()
	}

	wg.Wait()
	// Check that there were a reasonable number of hits.
	// It should be approx 50% but is random so we leave a large
	// margin of error and go for 10%.
	// If it's below this then something is definitely wrong.
	hitRatio := float64(hits) / float64(totalRequests)
	hitRatio = hitRatio * 100
	t.Log(fmt.Printf("Expected number of cache hits to be at least 10%% but was %s%%", fmt.Sprintf("%v", hitRatio)))
	if hitRatio < 10 {
		t.Error(AssertError)
		t.FailNow()
	}
}

/**
 * Check that a cache configured to not replace existing items does not do
 * it if an item with an existing key is added.
 */
func TestLruPutCache_DontReplace(t *testing.T) {
	// Create a cache. Use a size of two to rule out the case where the
	// second add removes the first by the LRU rules
	cb := CacheBuilder{CacheSize: 2, UpdateExisting: false}

	cache, err := NewLruPutCache(&cb)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = cache.Put(1, "test")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = cache.Put(1, "replacement")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	result, err := cache.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if fmt.Sprintf("%v", result) != "test" {
		t.Log("The existing value was not overwritten in the cache")
	} else {
		t.Log("The existing value was overwritten in the cache")
		t.FailNow()
	}
}

/**
 * Check that a cache configured to not replace existing items does so if an
 * item with an existing key is added.
 */
func TestLruPutCache_Replace(t *testing.T) {

	cb := CacheBuilder{CacheSize: 2, UpdateExisting: true}

	cache, err := NewLruPutCache(&cb)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = cache.Put(1, "test")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = cache.Put(1, "replacement")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	result, err := cache.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if fmt.Sprintf("%v", result) != "replacement" {
		t.Log("The existing value was not overwritten in the cache")
		t.FailNow()
	} else {
		t.Log("The existing value was overwritten in the cache")
	}
}