export type LogLevel = "TRACE" | "DEBUG" | "INFO" | "WARN" | "ERROR" | "FATAL";

export type LogEvent = {
  id: string;
  timestamp: string;
  service: string;
  environment: string;
  level: LogLevel;
  message: string;
  host?: string;
  trace_id?: string;
  metadata?: Record<string, unknown>;
  received_at: string;
};

export type SearchResponse = {
  logs: LogEvent[];
  total: number;
  limit: number;
  offset: number;
};

export type CountBucket = {
  name: string;
  value: number;
};

export type TimelineBucket = {
  time: string;
  logs: number;
};

export type AnalyticsSummary = {
  total: number;
  errors: number;
  warnings: number;
  error_rate: number;
  service_count: number;
  level_counts: CountBucket[];
  top_services: CountBucket[];
  top_errors: CountBucket[];
  timeline: TimelineBucket[];
};

export type SourceSummary = {
  service: string;
  environment: string;
  host_count: number;
  log_count: number;
  last_seen: string;
};

export type APIKey = {
  id: string;
  name: string;
  prefix: string;
  created_at: string;
  last_used_at?: string;
  revoked_at?: string;
  plaintext_key?: string;
};

export type RuntimeStats = {
  service: string;
  uptime_seconds: number;
  go_version: string;
  goroutines: number;
  allocated_bytes: number;
  heap_alloc_bytes: number;
  total_alloc_bytes: number;
  num_gc: number;
  stored_logs: number;
};

export type AuthResponse = {
  token: string;
  user: {
    id: string;
    project_id: string;
    email: string;
    created_at: string;
  };
};
