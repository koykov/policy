package policy

import (
	"sync"
	"sync/atomic"
)

type Lock struct {
	policy Policy
	mux    sync.Mutex
}

func (l *Lock) SetPolicy(new Policy) {
	atomic.StoreUint32((*uint32)(&l.policy), uint32(new))
}

func (l *Lock) GetPolicy() Policy {
	return Policy(atomic.LoadUint32((*uint32)(&l.policy)))
}

func (l *Lock) Lock() {
	if l.GetPolicy() == Locked {
		l.mux.Lock()
	}
}

func (l *Lock) Unlock() {
	if l.GetPolicy() == Locked {
		l.mux.Unlock()
	}
}
