package model

type RuntimeStats struct {
	Service         string `json:"service"`
	UptimeSeconds   int64  `json:"uptime_seconds"`
	GoVersion       string `json:"go_version"`
	Goroutines      int    `json:"goroutines"`
	AllocatedBytes  uint64 `json:"allocated_bytes"`
	HeapAllocBytes  uint64 `json:"heap_alloc_bytes"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
	NumGC           uint32 `json:"num_gc"`
	StoredLogs      int    `json:"stored_logs"`
}
