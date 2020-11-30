package policy

import (
	"sync"
	"sync/atomic"
)

type RWLock struct {
	policy Policy
	mux    sync.RWMutex

	lfcR, lcR int32
	lfcW, lcW int32
}

func (l *RWLock) SetPolicy(new Policy) {
	if new == Locked {
		l.waitL(new)
	}
	if new == LockFree {
		l.waitLF(new)
	}
}

func (l *RWLock) GetPolicy() Policy {
	return Policy(atomic.LoadUint32((*uint32)(&l.policy)))
}

func (l *RWLock) Lock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveL {
		l.mux.Lock()
		atomic.AddInt32(&l.lcW, 1)
		return
	}
	atomic.AddInt32(&l.lfcW, 1)
}

func (l *RWLock) Unlock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveLF {
		l.mux.Unlock()
		atomic.AddInt32(&l.lcW, -1)
		return
	}
	atomic.AddInt32(&l.lfcW, -1)
}

func (l *RWLock) RLock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveL {
		l.mux.RLock()
		atomic.AddInt32(&l.lcR, 1)
		return
	}
	atomic.AddInt32(&l.lfcR, 1)
}

func (l *RWLock) RUnlock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveLF {
		l.mux.RUnlock()
		atomic.AddInt32(&l.lcR, -1)
		return
	}
	atomic.AddInt32(&l.lfcR, -1)
}

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
