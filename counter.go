package rlutil

import "time"

type Counter interface {
	// Get returns the current count for the key and window
	Get(key string, window time.Time) (count int, err error) //nostyle:getters
	// Increment increments the count for the key and window
	Increment(key string, currWindow time.Time) error
}
