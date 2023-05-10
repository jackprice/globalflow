package globalflow

import "sync"

type Time int64

// LamportCLock contains an implementation of a Lamport clock.
type LamportCLock struct {
	// time is the current time.
	time Time

	// Mutex is a mutex for Vectors.
	// It must be held when writing to Vectors.
	mu sync.Mutex
}

// NewClock creates a new clock.
func NewClock() *LamportCLock {
	return &LamportCLock{}
}

// Get gets the current time.
func (clock *LamportCLock) Get() Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()

	clock.time++

	return clock.time
}

// Set updates the current time with a reference time from another node.
func (clock *LamportCLock) Set(time Time) {
	clock.mu.Lock()
	defer clock.mu.Unlock()

	if time > clock.time {
		clock.time = time
	}

	clock.time++
}

// VectorClock contains an implementation of a vector clock.
type VectorClock struct {
	// time contains a map of node IDs to times.
	time VectorTime

	// Mutex is a mutex for Vectors.
	mu sync.Mutex
}

type VectorTime map[string]Time

// NewVectorClock creates a new vector clock.
func NewVectorClock() *VectorClock {
	return &VectorClock{
		time: make(VectorTime),
	}
}

// Get gets the current time.
