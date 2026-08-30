package service

import (
	"context"
	"runtime"
	"time"

	"logmesh/internal/model"
)

type RuntimeService struct {
	logs      LogService
	startedAt time.Time
}

func NewRuntimeService(logs LogService) *RuntimeService {
	return &RuntimeService{
		logs:      logs,
		startedAt: time.Now().UTC(),
	}
}

func (s *RuntimeService) Stats(ctx context.Context) (model.RuntimeStats, error) {
	events, err := s.logs.Snapshot(ctx)
	if err != nil {
		return model.RuntimeStats{}, err
	}

	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)

	return model.RuntimeStats{
		Service:         "logmesh-api",
		UptimeSeconds:   int64(time.Since(s.startedAt).Seconds()),
		GoVersion:       runtime.Version(),
		Goroutines:      runtime.NumGoroutine(),
		AllocatedBytes:  memory.Alloc,
		HeapAllocBytes:  memory.HeapAlloc,
		TotalAllocBytes: memory.TotalAlloc,
		NumGC:           memory.NumGC,
		StoredLogs:      len(events),
	}, nil
}
