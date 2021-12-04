package policy

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

const (
	testStageCapRW = 1000
)

type testStageRW struct {
	lock RWLock
	data map[int32]int32
}

func (s *testStageRW) Fill(step int) {
	if step < 0 || step >= testStageCapRW {
		step = 1
	}
	s.lock.Lock()
	for i := 0; i < testStageCapRW; i += step {
		s.data[int32(i)] = rand.Int31()
	}
	s.lock.Unlock()
}

func (s *testStageRW) Write() {
	s.lock.Lock()
	s.data[rand.Int31n(testStageCapRW)] = rand.Int31()
	s.lock.Unlock()
}

func (s *testStageRW) Read() {
	s.lock.RLock()
	v, ok := s.data[rand.Int31n(testStageCapRW)]
	s.lock.RUnlock()
	_, _ = v, ok
}

func BenchmarkRWLockPolicy(b *testing.B) {
	b.Run("locked", func(b *testing.B) {
		stage := testStageRW{data: make(map[int32]int32, testStageCapRW)}
		b.ResetTimer()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if rand.Float64() < 0.5 {
					stage.Read()
				} else {
					stage.Write()
				}
			}
		})
	})
	b.Run("lock free", func(b *testing.B) {
		stage := testStageRW{data: make(map[int32]int32, testStageCapRW)}
		stage.Fill(2)
		stage.lock.SetPolicy(LockFree)

		b.ResetTimer()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				stage.Read()
			}
		})
	})
	b.Run("mixed", func(b *testing.B) {
		stage := testStageRW{data: make(map[int32]int32, testStageCapRW)}
		stage.Fill(testStageCapRW)
		stage.lock.SetPolicy(LockFree)

		var (
			wg    sync.WaitGroup
			done  = make([]chan struct{}, 100)
			state uint32
		)
		for i := 0; i < 100; i++ {
			wg.Add(1)
			done[i] = make(chan struct{}, 1)
			go func(done chan struct{}) {
				select {
				case <-done:
					wg.Done()
					return
				default:
					if atomic.LoadUint32(&state) == 0 {
						stage.Read()
					} else {
						if rand.Float64() < 0.5 {
							stage.Read()
						} else {
							stage.Write()
						}
					}
				}
			}(done[i])
		}

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if i%1e6 == 0 && i%2e6 != 0 {
				stage.lock.SetPolicy(Locked)
				atomic.StoreUint32(&state, 1)
			}
			if i%2e6 == 0 {
				atomic.StoreUint32(&state, 0)
				stage.lock.SetPolicy(LockFree)
			}
		}

		for i := 0; i < 100; i++ {
			done[i] <- struct{}{}
		}

		wg.Done()
	})
}
