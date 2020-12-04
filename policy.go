package policy

type Policy uint32

const (
	Locked Policy = iota
	LockFree
	// Transitive lock/lock-free policies.
	transitiveL
	transitiveLF
)
