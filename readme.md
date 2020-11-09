# Policy

Collection of behavior policies.

## Lock

A wrapper over `sync.Mutex`. Need for cases when your code writes in critical
code only occasionally. This policy may disable mutex lock/unlock during the
time when write is definitely not happens.

Usage:
```go
var (
    lock policy.Lock
    wg sync.WaitGroup
)
data := make(map[int32]bool, 1e6)

// Simultaneously write/read.
lock.SetPolicy(policy.Locked)
for i := 0; i < 1e6; i++ {
    // Writing/reading in concurrency.
    wg.Add(1)
    go func() {
        data[rand.Int31n(1e6)] = rand.Int31()
        wg.Done()
    }()
    wg.Add(1)
    go func() {
        _, _ = data[rand.Int31n(1e6)]
        wg.Done()
    }()
}
wg.Done()

// Only reading.
// Set lock policy as lock-free to suppress mutex lock/unlock calls and reduces
// `runtime.futex` pressure.
lock.SetPolicy(policy.LockFree)
for i := 0; i < 1e6; i++ {
    wg.Add(1)
    go func() {
        _, _ = data[rand.Int31n(1e6)]
        wg.Done()
    }()
}
// Restore locked state.
lock.SetPolicy(policy.Locked)
```

## RWLock.

A wrapper over `sync.RWMutex`. Work the same as `Lock` but allows you to protect
reading/writing separately.
