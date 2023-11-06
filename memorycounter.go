package rlutil

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

var _ Counter = (*MemoryCounter)(nil)

// MemoryCounter is a sliding window counter implemented with a TTL cache
type MemoryCounter struct {
	cache *ttlcache.Cache[string, *uint64]
	// windowLen is the length of the sliding window
	windowLen time.Duration
	// capacity is the maximum number of items to store in the cache
	capacity uint64
	// disableAutoDeleteExpired disables the automatic deletion of expired items
	disableAutoDeleteExpired bool
}

type MemoryCounterOption func(*MemoryCounter) error

// MemoryCounterWithCapacity sets the maximum number of items to store in the cache
func MemoryCounterWithCapacity(capacity uint64) MemoryCounterOption {
	return func(c *MemoryCounter) error {
		c.capacity = capacity
		return nil
	}
}

// MemoryCounterDisableAutoDeleteExpired disables the automatic deletion of expired items
func MemoryCounterDisableAutoDeleteExpired() MemoryCounterOption {
	return func(c *MemoryCounter) error {
		c.disableAutoDeleteExpired = true
		return nil
	}
}

// NewMemoryCounter creates a new MemoryCounter
func NewMemoryCounter(windowLen time.Duration, opts ...MemoryCounterOption) *MemoryCounter {
	c := &MemoryCounter{}
	for _, opt := range opts {
		opt(c)
	}
	ttlOpts := []ttlcache.Option[string, *uint64]{
		ttlcache.WithTTL[string, *uint64](windowLen * 2),
	}
	if c.capacity > 0 {
		ttlOpts = append(ttlOpts, ttlcache.WithCapacity[string, *uint64](c.capacity))
	}
	cache := ttlcache.New[string, *uint64](ttlOpts...)
	c.cache = cache
	if !c.disableAutoDeleteExpired {
		go cache.Start()
	}
	return c
}

// Get returns the count for the given key and window
func (c *MemoryCounter) Get(key string, window time.Time) (count int, err error) {
	key = generateKey(key, window)
	i := c.cache.Get(key)
	if i == nil {
		return 0, nil
	}
	return int(*i.Value()), nil
}

// Increment increments the count for the given key and window
func (c *MemoryCounter) Increment(key string, currWindow time.Time) error {
	key = generateKey(key, currWindow)
	zero := uint64(0)
	i, _ := c.cache.GetOrSet(key, &zero)
	atomic.AddUint64(i.Value(), 1)
	return nil
}

func generateKey(key string, window time.Time) string {
	return fmt.Sprintf("%s-%d", key, window.Unix())
}

// DeleteExpired deletes expired items from the cache
func (c *MemoryCounter) DeleteExpired() {
	c.cache.DeleteExpired()
}
