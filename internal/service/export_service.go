package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"

	"logmesh/internal/model"
)

type ExportService struct {
	logs LogService
}

func NewExportService(logs LogService) *ExportService {
	return &ExportService{logs: logs}
}

func (s *ExportService) CSV(ctx context.Context, query model.SearchLogsQuery) ([]byte, error) {
	result, err := s.logs.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	if err := writer.Write([]string{"id", "timestamp", "level", "service", "environment", "host", "trace_id", "message"}); err != nil {
		return nil, err
	}

	for _, event := range result.Logs {
		if err := writer.Write([]string{
			event.ID,
			event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			string(event.Level),
			event.Service,
			event.Environment,
			event.Host,
			event.TraceID,
			event.Message,
		}); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return []byte("# total," + strconv.Itoa(result.Total) + "\n" + buffer.String()), nil
}
