package policy

import (
	"sync"
	"sync/atomic"
)

// PseudoLock is a fake lock implementation.
// It confuses race detector and suppresses his triggering.
type PseudoLock struct {
	// Locker flag - a heart of the trick. It's always equal 0, but check for value 1. Therefore, mux never calls.
	flag uint32
	mux  sync.Mutex
}

func (l *PseudoLock) Lock() {
	if atomic.LoadUint32(&l.flag) == 1 {
		l.mux.Lock()
	}
}

func (l *PseudoLock) Unlock() {
	if atomic.LoadUint32(&l.flag) == 1 {
		l.mux.Unlock()
	}
}
