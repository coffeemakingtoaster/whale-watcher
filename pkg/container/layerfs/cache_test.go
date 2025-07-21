package layerfs_test

import (
	"testing"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container/layerfs"
)

func TestPut(t *testing.T) {
	cache := layerfs.NewLRUCache[int](3)
	items := []string{"first", "second", "third"}
	for i := range items {
		cache.Put(items[i], i)

		if cache.GetSize() != i+1 {
			t.Errorf("Size mismatch: Expected %d Got %d", i+1, cache.GetSize())
		}
	}
}

func TestPutUpdate(t *testing.T) {
	cache := layerfs.NewLRUCache[int](3)
	items := []string{"first", "second", "third"}
	for i := range items {
		cache.Put(items[i], i)
	}

	cache.Put("first", 42)

	if *cache.Get("first") != 42 {
		t.Errorf("Value return mismatch: Expected 0 Got %v", *cache.Get("first"))
	}

	if cache.Get("second") == nil {
		t.Error("Evict mismatch: Expected no evict but got evict")
	}
}

func TestPutEvict(t *testing.T) {
	cache := layerfs.NewLRUCache[int](3)
	items := []string{"first", "second", "third"}
	for i := range items {
		cache.Put(items[i], i)
	}

	if *cache.Get("first") != 0 {
		t.Errorf("Value return mismatch: Expected 0 Got %v", *cache.Get("first"))
	}

	cache.Put("too much", 100)

	if cache.Get("second") != nil {
		t.Errorf("Evict mismatch: Expected nil but got value Got %v", cache.Get("second"))
	}
}

func TestUpdateEvict(t *testing.T) {
	cache := layerfs.NewLRUCache[int](3)
	items := []string{"first", "second", "third"}
	for i := range items {
		cache.Put(items[i], i)
	}

	cache.Put("first", 1)
	cache.Put("second", 42)

	if *cache.Get("first") != 1 {
		t.Errorf("Value return mismatch: Expected 0 Got %v", *cache.Get("first"))
	}

	cache.Put("too much", 100)

	if cache.Get("third") != nil {
		t.Errorf("Evict mismatch: Expected nil but got value Got %v", cache.Get("second"))
	}

	if *cache.Get("second") != 42 {
		t.Errorf("Value return mismatch: Expected 0 Got %v", *cache.Get("first"))
	}
}
