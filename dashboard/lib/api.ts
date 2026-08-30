import { apiBase } from "./constants";
import type { APIKey, AnalyticsSummary, AuthResponse, LogEvent, RuntimeStats, SourceSummary } from "./types";

export async function fetchAnalytics() {
  const response = await fetch(`${apiBase}/v1/analytics`, { cache: "no-store", headers: authHeaders() });
  if (!response.ok) throw new Error("failed to fetch analytics");
  return (await response.json()) as AnalyticsSummary;
}

export async function fetchSources() {
  const response = await fetch(`${apiBase}/v1/sources`, { cache: "no-store", headers: authHeaders() });
  if (!response.ok) throw new Error("failed to fetch sources");
  const payload = (await response.json()) as { sources: SourceSummary[] };
  return payload.sources;
}

export async function fetchAPIKeys() {
  const response = await fetch(`${apiBase}/v1/api-keys`, { cache: "no-store", headers: authHeaders() });
  if (!response.ok) throw new Error("failed to fetch api keys");
  const payload = (await response.json()) as { api_keys: APIKey[] };
  return payload.api_keys;
}

export async function createAPIKey(name: string) {
  const response = await fetch(`${apiBase}/v1/api-keys`, {
    method: "POST",
    headers: { ...authHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify({ name })
  });
  if (!response.ok) throw new Error("failed to create api key");
  return (await response.json()) as APIKey;
}

export async function revokeAPIKey(id: string) {
  const response = await fetch(`${apiBase}/v1/api-keys/${id}`, {
    method: "DELETE",
    headers: authHeaders()
  });
  if (!response.ok) throw new Error("failed to revoke api key");
}

export async function parseTextLog(input: {
  service: string;
  environment: string;
  host: string;
  trace_id: string;
  line: string;
}) {
  const response = await fetch(`${apiBase}/v1/logs/parse`, {
    method: "POST",
    headers: { ...authHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify(input)
  });
  if (!response.ok) throw new Error("failed to parse log");
  return (await response.json()) as LogEvent;
}

export async function bulkIngestLogs(logs: Array<Omit<LogEvent, "id" | "timestamp" | "received_at">>) {
  const response = await fetch(`${apiBase}/v1/logs/bulk`, {
    method: "POST",
    headers: { ...authHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify({ logs })
  });
  if (!response.ok) throw new Error("failed to ingest bulk logs");
  return (await response.json()) as { accepted: number; logs: LogEvent[] };
}

export async function fetchRuntime() {
  const response = await fetch(`${apiBase}/v1/runtime`, { cache: "no-store", headers: authHeaders() });
  if (!response.ok) throw new Error("failed to fetch runtime stats");
  return (await response.json()) as RuntimeStats;
}

export function logExportURL() {
  return `${apiBase}/v1/logs/export?limit=500`;
}

export async function registerUser(email: string, password: string) {
  return authRequest("/v1/auth/register", email, password);
}

export function authHeaders() {
  if (typeof window === "undefined") {
    return {};
  }
  const token = window.localStorage.getItem("logmesh_token");
  return token ? { Authorization: `Bearer ${token}` } : {};
}

export async function loginUser(email: string, password: string) {
  return authRequest("/v1/auth/login", email, password);
}

async function authRequest(path: string, email: string, password: string) {
  const response = await fetch(`${apiBase}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password })
  });
  if (!response.ok) throw new Error("authentication failed");
  return (await response.json()) as AuthResponse;
}
