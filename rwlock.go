package policy

import (
	"sync"
	"sync/atomic"
)

type RWLock struct {
	policy Policy
	mux    sync.RWMutex
}

func (l *RWLock) SetPolicy(new Policy) {
	atomic.StoreUint32((*uint32)(&l.policy), uint32(new))
}

func (l *RWLock) GetPolicy() Policy {
	return Policy(atomic.LoadUint32((*uint32)(&l.policy)))
}

func (l *RWLock) Lock() {
	if l.GetPolicy() == Locked {
		l.mux.Lock()
	}
}

func (l *RWLock) Unlock() {
	if l.GetPolicy() == Locked {
		l.mux.Unlock()
	}
}

func (l *RWLock) RLock() {
	if l.GetPolicy() == Locked {
		l.mux.RLock()
	}
}

func (l *RWLock) RUnlock() {
	if l.GetPolicy() == Locked {
		l.mux.RUnlock()
	}
}
