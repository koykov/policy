package policy

type Policy uint32

const (
	Locked Policy = iota
	LockFree
)
