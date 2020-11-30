package policy

import (
	"sync"
	"sync/atomic"
)

type Lock struct {
	policy  Policy
	mux     sync.Mutex
	lfc, lc int32
}

func (l *Lock) SetPolicy(new Policy) {
	if new == Locked {
		l.waitLF(new)
	}
	if new == LockFree {
		l.waitL(new)
	}
}

func (l *Lock) GetPolicy() Policy {
	return Policy(atomic.LoadUint32((*uint32)(&l.policy)))
}

func (l *Lock) Lock() {
	policy := l.GetPolicy()
	if policy == Locked || policy == transitiveL {
		l.mux.Lock()
		atomic.AddInt32(&l.lc, 1)
		return
	}
	atomic.AddInt32(&l.lfc, 1)
}

func (l *Lock) Unlock() {
	if l.GetPolicy() == Locked {
		l.mux.Unlock()
		atomic.AddInt32(&l.lc, -1)
		return
	}
	atomic.AddInt32(&l.lfc, -1)
}

func (l *Lock) waitLF(final Policy) {
	l.mux.Lock()
	if l.GetPolicy() == final {
		l.mux.Unlock()
		return
	}
	atomic.StoreUint32((*uint32)(&l.policy), uint32(transitiveL))
	for atomic.LoadInt32(&l.lfc) > 0 {
	}
	atomic.StoreUint32((*uint32)(&l.policy), uint32(final))
	l.mux.Unlock()
}

func (l *Lock) waitL(final Policy) {
	l.mux.Lock()
	if l.GetPolicy() == final {
		l.mux.Unlock()
		return
	}
	atomic.StoreUint32((*uint32)(&l.policy), uint32(transitiveLF))
	for atomic.LoadInt32(&l.lc) > 0 {
	}
	atomic.StoreUint32((*uint32)(&l.policy), uint32(final))
	l.mux.Unlock()
}
