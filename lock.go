package policy

import (
	"sync"
	"sync/atomic"
)

// Lock policy.
type Lock struct {
	policy Policy
	mux    sync.Mutex
	// Lock-free and lock counters.
	lfc, lc int32
}

// Set new policy.
//
// Use it to enable Locked policy before critical sections and switch back to LockFree afterward.
func (l *Lock) SetPolicy(new Policy) {
	if new == Locked {
		l.waitLF(new)
	}
	if new == LockFree {
		l.waitL(new)
	}
}

// Get current policy.
func (l *Lock) GetPolicy() Policy {
	return Policy(atomic.LoadUint32((*uint32)(&l.policy)))
}

// Lock internal mutex according policy.
func (l *Lock) Lock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveL {
		// Lock mutex in Locked and transitive to lock states.
		l.mux.Lock()
		// Increase locked counter.
		atomic.AddInt32(&l.lc, 1)
		return
	}
	// Increase lock-free counter.
	atomic.AddInt32(&l.lfc, 1)
}

// Unlock internal mutex according policy.
func (l *Lock) Unlock() {
	if policy := l.GetPolicy(); policy == Locked || policy == transitiveLF {
		// Unlock mutex in Locked and transitive to lock-free states.
		l.mux.Unlock()
		// Decrease locked counter.
		atomic.AddInt32(&l.lc, -1)
		return
	}
	// Decrease lock-free counter.
	atomic.AddInt32(&l.lfc, -1)
}

// Wait finishing of all lock-free routines.
func (l *Lock) waitLF(final Policy) {
	// Protect waiting.
	l.mux.Lock()
	// Check if policy already changed in parallel routine.
	if l.GetPolicy() == final {
		// Unlock and exit.
		l.mux.Unlock()
		return
	}
	// Set the transitive lock state.
	atomic.StoreUint32((*uint32)(&l.policy), uint32(transitiveL))
	// Wait while all lock-free routines finished work.
	for atomic.LoadInt32(&l.lfc) > 0 {
	}
	// Set the target state.
	atomic.StoreUint32((*uint32)(&l.policy), uint32(final))
	l.mux.Unlock()
}

func (l *Lock) waitL(final Policy) {
	// Protect waiting.
	l.mux.Lock()
	// Check if policy already changed in parallel routine.
	if l.GetPolicy() == final {
		// Unlock and exit.
		l.mux.Unlock()
		return
	}
	// Set the transitive lock-free state.
	atomic.StoreUint32((*uint32)(&l.policy), uint32(transitiveLF))
	// Wait while all locked routines finished work.
	for atomic.LoadInt32(&l.lc) > 0 {
	}
	// Set the target state.
	atomic.StoreUint32((*uint32)(&l.policy), uint32(final))
	l.mux.Unlock()
}
