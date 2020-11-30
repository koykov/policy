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

func BenchmarkLockPolicyLocked(b *testing.B) {
	stage := testStage{data: make(map[int32]int32, testStageCap)}
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
}

func BenchmarkLockPolicyLockFree(b *testing.B) {
	stage := testStage{data: make(map[int32]int32, testStageCap)}
	stage.Fill(2)
	stage.lock.SetPolicy(LockFree)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stage.Read()
		}
	})
}

func BenchmarkLockPolicyMixed(b *testing.B) {
	stage := testStage{data: make(map[int32]int32, testStageCap)}
	stage.Fill(testStageCap)
	stage.lock.SetPolicy(LockFree)

	var (
		wg    sync.WaitGroup
		done  = make([]chan bool, 100)
		state uint32
	)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		done[i] = make(chan bool, 1)
		go func(done chan bool) {
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
		done[i] <- true
	}

	wg.Done()
}
