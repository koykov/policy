package policy

import (
	"sync"
	"sync/atomic"
)

// RWLock is a read/write lock policy implementation.
type RWLock struct {
	policy Policy
	mux    sync.RWMutex
	// Lock-free/lock read counters.
	lfcR, lcR int32
	// Lock-free/lock write counters.
	lfcW, lcW int32
}

// SetPolicy sets lock's policy.
func (l *RWLock) SetPolicy(new Policy) {
	if new == Locked {
		l.waitLF(new)
	}
	if new == LockFree {
		l.waitL(new)
	}
}

// GetPolicy returns current policy.
func (l *RWLock) GetPolicy() Policy {
	return Policy(atomic.LoadUint32((*uint32)(&l.policy)))
}

// Lock locks RW using internal mutex according policy.
func (l *RWLock) Lock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveL {
		// Lock mutex in Locked or transitive to lock states.
		l.mux.Lock()
		// Increase locked write counter.
		atomic.AddInt32(&l.lcW, 1)
		return
	}
	// Increase lock-free write counter.
	atomic.AddInt32(&l.lfcW, 1)
}

// Unlock unlocks RW using internal mutex according policy.
func (l *RWLock) Unlock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveLF {
		// Decrease locked write counter.
		atomic.AddInt32(&l.lcW, -1)
		// Unlock mutex in Locked or transitive to lock-free states.
		l.mux.Unlock()
		return
	}
	// Decrease lock-free write counter.
	atomic.AddInt32(&l.lfcW, -1)
}

// RLock locks read using internal mutex according policy.
func (l *RWLock) RLock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveL {
		// Lock mutex in Locked or transitive to lock states.
		l.mux.RLock()
		// Increase locked read counter.
		atomic.AddInt32(&l.lcR, 1)
		return
	}
	// Increase lock-free read counter.
	atomic.AddInt32(&l.lfcR, 1)
}

// RUnlock unlocks read using internal mutex according policy.
func (l *RWLock) RUnlock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveLF {
		// Decrease locked read counter.
		atomic.AddInt32(&l.lcR, -1)
		// Unlock mutex in Locked or transitive to lock-free states.
		l.mux.RUnlock()
		return
	}
	// Decrease lock-free read counter.
	atomic.AddInt32(&l.lfcR, -1)
}

// Wait finishing of all lock-free routines.
func (l *RWLock) waitLF(final Policy) {
	l.mux.Lock()
	if l.GetPolicy() == final {
		l.mux.Unlock()
		return
	}
	atomic.StoreUint32((*uint32)(&l.policy), uint32(transitiveL))
	for atomic.LoadInt32(&l.lfcR) > 0 && atomic.LoadInt32(&l.lfcW) > 0 {
	}
	atomic.StoreUint32((*uint32)(&l.policy), uint32(final))
	l.mux.Unlock()
}

// Wait finishing of all locked routines.
func (l *RWLock) waitL(final Policy) {
	l.mux.Lock()
	if l.GetPolicy() == final {
		l.mux.Unlock()
		return
	}
	atomic.StoreUint32((*uint32)(&l.policy), uint32(transitiveLF))
	for atomic.LoadInt32(&l.lcR) > 0 && atomic.LoadInt32(&l.lcW) > 0 {
	}
	atomic.StoreUint32((*uint32)(&l.policy), uint32(final))
	l.mux.Unlock()
}
