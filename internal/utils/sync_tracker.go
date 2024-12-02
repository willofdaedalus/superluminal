package utils

import (
	"sync/atomic"
)

type SyncTracker struct {
	activeWrites int64
	activeReads  int64
}

func NewSyncTracker() *SyncTracker {
	return &SyncTracker{
		activeWrites: 0,
		activeReads:  0,
	}
}

// write tracking methods
func (s *SyncTracker) IncrementWrite() {
	atomic.AddInt64(&s.activeWrites, 1)
}

func (s *SyncTracker) DecrementWrite() {
	atomic.AddInt64(&s.activeWrites, -1)
}

func (s *SyncTracker) HasWrites() bool {
	return atomic.LoadInt64(&s.activeWrites) > 0
}

// read tracking methods
func (s *SyncTracker) IncrementRead() {
	atomic.AddInt64(&s.activeReads, 1)
}

func (s *SyncTracker) DecrementRead() {
	atomic.AddInt64(&s.activeReads, -1)
}

func (s *SyncTracker) HasReads() bool {
	return atomic.LoadInt64(&s.activeReads) > 0
}

// combined tracking methods
// probably shouldn't be here but remember to check that there's nothing
// in both server and client's respective queues
func (s *SyncTracker) AnyActionInProgress() bool {
	return s.HasWrites() || s.HasReads()
}
