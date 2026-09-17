package metrics

import (
	"sync"
	"sync/atomic"
)

type Metrics struct {
	mu             sync.Mutex
	totalRequests  int64
	requestsByPath map[string]int64
	totalLatencyMs float64
}

func NewMetrics() *Metrics {
	return &Metrics{requestsByPath: make(map[string]int64)}
}

func (m *Metrics) Record(path string, latencyMs float64) {
	atomic.AddInt64(&m.totalRequests, 1)
	m.mu.Lock()
	m.requestsByPath[path]++
	m.totalLatencyMs += latencyMs
	m.mu.Unlock()
}

func (m *Metrics) Snapshot() (total int64, byPath map[string]int64, avgLatency float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	total = atomic.LoadInt64(&m.totalRequests)
	byPath = make(map[string]int64, len(m.requestsByPath))
	for k, v := range m.requestsByPath {
		byPath[k] = v
	}
	if total > 0 {
		avgLatency = m.totalLatencyMs / float64(total)
	}
	return
}
