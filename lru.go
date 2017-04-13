// Copyright 2016 zxfonline@sina.com. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Package lru implements an LRU cache.

package lru

import "container/list"

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	// OnEvicted optionally specificies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key Key, value interface{})

	Ll    *list.List
	Cache map[interface{}]*list.Element
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

type entry struct {
	key   Key
	value interface{}
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		Ll:         list.New(),
		Cache:      make(map[interface{}]*list.Element),
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key Key, value interface{}) {
	if c.Cache == nil {
		c.Cache = make(map[interface{}]*list.Element)
		c.Ll = list.New()
	}
	if ee, ok := c.Cache[key]; ok {
		c.Ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.Ll.PushFront(&entry{key, value})
	c.Cache[key] = ele
	if c.MaxEntries != 0 && c.Ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.Cache == nil {
		return
	}
	if ele, hit := c.Cache[key]; hit {
		c.Ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key Key) {
	if c.Cache == nil {
		return
	}
	if ele, hit := c.Cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() Key {
	if c.Cache == nil {
		return nil
	}
	ele := c.Ll.Back()
	if ele != nil {
		c.removeElement(ele)
		return ele.Value.(*entry).key
	}
	return nil
}

func (c *Cache) removeElement(e *list.Element) {
	c.Ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.Cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	if c.Cache == nil {
		return 0
	}
	return c.Ll.Len()
}

// Foreach foreach the oldest item from the cache.
//fn return args
//arg1:if true break foreach,or continue foreach
func (c *Cache) Foreach(fn func(Key, interface{}) bool) {
	if c.Cache == nil {
		return
	}
	var ret bool
	for ele := c.Ll.Back(); ele != nil; ele = ele.Prev() {
		entry := ele.Value.(*entry)
		if ret = fn(entry.key, entry.value); ret {
			break
		}
	}
}

// Foreach foreach the oldest item from the cache.
//fn return args
//arg1:true break foreach,or continue foreach.
//arg2:true delete element from the cache.
func (c *Cache) RemoveForeach(fn func(Key, interface{}) (bool, bool)) {
	if c.Cache == nil {
		return
	}
	var remove, ret bool
	for ele := c.Ll.Back(); ele != nil; {
		entry := ele.Value.(*entry)
		oldEle := ele
		ele = ele.Prev()
		ret, remove = fn(entry.key, entry.value)
		if ret {
			break
		}
		if remove {
			c.removeElement(oldEle)
		}
	}
}
