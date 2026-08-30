package model

type CountBucket struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type TimelineBucket struct {
	Time string `json:"time"`
	Logs int    `json:"logs"`
}

type AnalyticsSummary struct {
	Total        int              `json:"total"`
	Errors       int              `json:"errors"`
	Warnings     int              `json:"warnings"`
	ErrorRate    float64          `json:"error_rate"`
	ServiceCount int              `json:"service_count"`
	LevelCounts  []CountBucket    `json:"level_counts"`
	TopServices  []CountBucket    `json:"top_services"`
	TopErrors    []CountBucket    `json:"top_errors"`
	Timeline     []TimelineBucket `json:"timeline"`
}

type SourceSummary struct {
	Service     string `json:"service"`
	Environment string `json:"environment"`
	HostCount   int    `json:"host_count"`
	LogCount    int    `json:"log_count"`
	LastSeen    string `json:"last_seen"`
}
