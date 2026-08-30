package model

type BulkIngestLogRequest struct {
	Logs []IngestLogRequest `json:"logs"`
}

type BulkIngestLogResponse struct {
	Accepted int        `json:"accepted"`
	Logs     []LogEvent `json:"logs"`
}

type ParseTextLogRequest struct {
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Host        string `json:"host"`
	TraceID     string `json:"trace_id"`
	Line        string `json:"line"`
}
