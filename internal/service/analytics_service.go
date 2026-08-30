package service

import (
	"context"
	"sort"
	"time"

	"logmesh/internal/model"
)

type AnalyticsService struct {
	logs LogService
}

func NewAnalyticsService(logs LogService) *AnalyticsService {
	return &AnalyticsService{logs: logs}
}

func (s *AnalyticsService) Summary(ctx context.Context) (model.AnalyticsSummary, error) {
	events, err := s.logs.Snapshot(ctx)
	if err != nil {
		return model.AnalyticsSummary{}, err
	}

	levelCounts := map[string]int{}
	serviceCounts := map[string]int{}
	errorCounts := map[string]int{}
	timelineCounts := map[string]int{}
	errors := 0
	warnings := 0

	for _, event := range events {
		levelCounts[string(event.Level)]++
		serviceCounts[event.Service]++

		switch event.Level {
		case model.LevelError, model.LevelFatal:
			errors++
			errorCounts[event.Message]++
		case model.LevelWarn:
			warnings++
		}

		bucket := event.Timestamp.UTC().Truncate(time.Minute).Format("15:04")
		timelineCounts[bucket]++
	}

	errorRate := 0.0
	if len(events) > 0 {
		errorRate = (float64(errors) / float64(len(events))) * 100
	}

	return model.AnalyticsSummary{
		Total:        len(events),
		Errors:       errors,
		Warnings:     warnings,
		ErrorRate:    errorRate,
		ServiceCount: len(serviceCounts),
		LevelCounts:  fixedLevelBuckets(levelCounts),
		TopServices:  topBuckets(serviceCounts, 6),
		TopErrors:    topBuckets(errorCounts, 6),
		Timeline:     timelineBuckets(timelineCounts),
	}, nil
}

func (s *AnalyticsService) Sources(ctx context.Context) ([]model.SourceSummary, error) {
	events, err := s.logs.Snapshot(ctx)
	if err != nil {
		return nil, err
	}

	type aggregate struct {
		hosts    map[string]struct{}
		logCount int
		lastSeen time.Time
	}

	groups := map[string]*aggregate{}
	for _, event := range events {
		key := event.Service + "\x00" + event.Environment
		group := groups[key]
		if group == nil {
			group = &aggregate{hosts: map[string]struct{}{}}
			groups[key] = group
		}
		if event.Host != "" {
			group.hosts[event.Host] = struct{}{}
		}
		group.logCount++
		if event.Timestamp.After(group.lastSeen) {
			group.lastSeen = event.Timestamp
		}
	}

	sources := make([]model.SourceSummary, 0, len(groups))
	for key, group := range groups {
		parts := splitSourceKey(key)
		sources = append(sources, model.SourceSummary{
			Service:     parts[0],
			Environment: parts[1],
			HostCount:   len(group.hosts),
			LogCount:    group.logCount,
			LastSeen:    group.lastSeen.UTC().Format(time.RFC3339),
		})
	}

	sort.Slice(sources, func(i, j int) bool {
		return sources[i].LogCount > sources[j].LogCount
	})

	return sources, nil
}

func fixedLevelBuckets(counts map[string]int) []model.CountBucket {
	levels := []model.LogLevel{
		model.LevelTrace,
		model.LevelDebug,
		model.LevelInfo,
		model.LevelWarn,
		model.LevelError,
		model.LevelFatal,
	}

	buckets := make([]model.CountBucket, 0, len(levels))
	for _, level := range levels {
		buckets = append(buckets, model.CountBucket{Name: string(level), Value: counts[string(level)]})
	}
	return buckets
}

func topBuckets(counts map[string]int, limit int) []model.CountBucket {
	buckets := make([]model.CountBucket, 0, len(counts))
	for name, value := range counts {
		buckets = append(buckets, model.CountBucket{Name: name, Value: value})
	}

	sort.Slice(buckets, func(i, j int) bool {
		if buckets[i].Value == buckets[j].Value {
			return buckets[i].Name < buckets[j].Name
		}
		return buckets[i].Value > buckets[j].Value
	})

	if len(buckets) > limit {
		return buckets[:limit]
	}
	return buckets
}

func timelineBuckets(counts map[string]int) []model.TimelineBucket {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	buckets := make([]model.TimelineBucket, 0, len(keys))
	for _, key := range keys {
		buckets = append(buckets, model.TimelineBucket{Time: key, Logs: counts[key]})
	}
	return buckets
}

func splitSourceKey(key string) [2]string {
	for i := range key {
		if key[i] == 0 {
			return [2]string{key[:i], key[i+1:]}
		}
	}
	return [2]string{key, ""}
}
