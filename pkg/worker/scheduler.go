package worker

import (
	"sync"
)

type Scheduler struct {
	workers map[string]Worker
	mutex   sync.RWMutex
}

func (s *Scheduler) Init() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.workers = make(map[string]Worker)
}

func (s *Scheduler) Register(name string, w Worker) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.workers[name] = w
}

func (s *Scheduler) Unregister(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.workers, name)
}

func (s *Scheduler) StartWorker() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, w := range s.workers {
		w.Start()
	}
}

func (s *Scheduler) StopWorker() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, w := range s.workers {
		w.Stop()
	}
}
