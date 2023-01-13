package policy

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

const (
	testStageCap = 1000
)

type testStage struct {
	lock Lock
	data map[int32]int32
}

func (s *testStage) Fill(step int) {
	if step < 0 || step >= testStageCap {
		step = 1
	}
	s.lock.Lock()
	for i := 0; i < testStageCap; i += step {
		s.data[int32(i)] = rand.Int31()
	}
	s.lock.Unlock()
}

func (s *testStage) Write() {
	s.lock.Lock()
	s.data[rand.Int31n(testStageCap)] = rand.Int31()
	s.lock.Unlock()
}

func (s *testStage) Read() {
	s.lock.Lock()
	v, ok := s.data[rand.Int31n(testStageCap)]
	s.lock.Unlock()
	_, _ = v, ok
}

func BenchmarkLockPolicy(b *testing.B) {
	b.Run("locked", func(b *testing.B) {
		stage := testStage{data: make(map[int32]int32, testStageCap)}
		b.ResetTimer()
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
		stage := testStage{data: make(map[int32]int32, testStageCap)}
		stage.Fill(2)
		stage.lock.SetPolicy(LockFree)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				stage.Read()
			}
		})
	})
	b.Run("mixed", func(b *testing.B) {
		const workers = 1

		stage := testStage{data: make(map[int32]int32, testStageCap)}
		stage.Fill(testStageCap)
		stage.lock.SetPolicy(LockFree)

		var (
			wg    sync.WaitGroup
			done  = make([]chan struct{}, workers)
			state uint32
		)
		for i := 0; i < workers; i++ {
			wg.Add(1)
			done[i] = make(chan struct{}, 1)
			go func(done chan struct{}) {
				for {
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
				}
			}(done[i])
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%1e6 == 0 && i%2e6 != 0 {
				print(atomic.LoadUint32(&state))
				stage.lock.SetPolicy(Locked)
				atomic.StoreUint32(&state, 1)
				println(atomic.LoadUint32(&state))
			}
			if i%2e6 == 0 {
				print(atomic.LoadUint32(&state))
				atomic.StoreUint32(&state, 0)
				stage.lock.SetPolicy(LockFree)
				println(atomic.LoadUint32(&state))
			}
		}

		for i := 0; i < workers; i++ {
			done[i] <- struct{}{}
		}

		wg.Wait()
	})
}
