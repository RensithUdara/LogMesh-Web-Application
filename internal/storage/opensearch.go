package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"

	"logmesh/internal/model"
)

type OpenSearchStore struct {
	client *opensearch.Client
	index  string
}

func NewOpenSearchStore(url string) (*OpenSearchStore, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, nil
	}

	client, err := opensearch.NewClient(opensearch.Config{Addresses: []string{url}})
	if err != nil {
		return nil, err
	}

	return &OpenSearchStore{client: client, index: "logs-*"}, nil
}

func (s *OpenSearchStore) Index(ctx context.Context, event model.LogEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	index := fmt.Sprintf("logs-%s", event.Timestamp.UTC().Format("2006.01.02"))
	req := opensearchapi.IndexRequest{
		Index:      index,
		DocumentID: event.ID,
		Body:       bytes.NewReader(body),
		Refresh:    "false",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		payload, _ := io.ReadAll(res.Body)
		return fmt.Errorf("opensearch index failed: %s %s", res.Status(), strings.TrimSpace(string(payload)))
	}
	return nil
}

func (s *OpenSearchStore) BulkIndex(ctx context.Context, events []model.LogEvent) error {
	if len(events) == 0 {
		return nil
	}

	var body bytes.Buffer
	for _, event := range events {
		index := fmt.Sprintf("logs-%s", event.Timestamp.UTC().Format("2006.01.02"))
		meta := map[string]any{"index": map[string]any{"_index": index, "_id": event.ID}}
		metaBytes, _ := json.Marshal(meta)
		eventBytes, err := json.Marshal(event)
		if err != nil {
			return err
		}
		body.Write(metaBytes)
		body.WriteByte('\n')
		body.Write(eventBytes)
		body.WriteByte('\n')
	}

	req := opensearchapi.BulkRequest{Body: &body}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		payload, _ := io.ReadAll(res.Body)
		return fmt.Errorf("opensearch bulk failed: %s %s", res.Status(), strings.TrimSpace(string(payload)))
	}
	return nil
}

func (s *OpenSearchStore) Search(ctx context.Context, query model.SearchLogsQuery) (model.SearchLogsResult, error) {
	body, err := buildSearchBody(query)
	if err != nil {
		return model.SearchLogsResult{}, err
	}

	req := opensearchapi.SearchRequest{
		Index: []string{s.index},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return model.SearchLogsResult{}, err
	}
	defer res.Body.Close()

	if res.IsError() {
		payload, _ := io.ReadAll(res.Body)
		return model.SearchLogsResult{}, fmt.Errorf("opensearch search failed: %s %s", res.Status(), strings.TrimSpace(string(payload)))
	}

	var payload struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source model.LogEvent `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return model.SearchLogsResult{}, err
	}

	logs := make([]model.LogEvent, 0, len(payload.Hits.Hits))
	for _, hit := range payload.Hits.Hits {
		logs = append(logs, hit.Source)
	}

	return model.SearchLogsResult{
		Logs:   logs,
		Total:  payload.Hits.Total.Value,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func buildSearchBody(query model.SearchLogsQuery) ([]byte, error) {
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 500 {
		query.Limit = 500
	}

	filters := make([]map[string]any, 0)
	terms := map[string]string{
		"project_id.keyword":  query.ProjectID,
		"service.keyword":     query.Service,
		"environment.keyword": query.Environment,
		"level.keyword":       string(query.Level),
		"trace_id.keyword":    query.TraceID,
		"host.keyword":        query.Host,
	}
	for field, value := range terms {
		if value != "" {
			filters = append(filters, map[string]any{"term": map[string]any{field: value}})
		}
	}
	if query.From != nil || query.To != nil {
		rng := map[string]any{}
		if query.From != nil {
			rng["gte"] = query.From.UTC().Format("2006-01-02T15:04:05Z07:00")
		}
		if query.To != nil {
			rng["lte"] = query.To.UTC().Format("2006-01-02T15:04:05Z07:00")
		}
		filters = append(filters, map[string]any{"range": map[string]any{"timestamp": rng}})
	}

	must := make([]map[string]any, 0)
	if query.Search != "" {
		must = append(must, map[string]any{
			"match": map[string]any{"message": query.Search},
		})
	}

	body := map[string]any{
		"from": query.Offset,
		"size": query.Limit,
		"sort": []map[string]any{{"timestamp": map[string]any{"order": "desc"}}},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
				"must":   must,
			},
		},
	}
	return json.Marshal(body)
}
