import { apiBase } from "./constants";
import type { APIKey, AnalyticsSummary, SourceSummary } from "./types";

export async function fetchAnalytics() {
  const response = await fetch(`${apiBase}/v1/analytics`, { cache: "no-store" });
  if (!response.ok) throw new Error("failed to fetch analytics");
  return (await response.json()) as AnalyticsSummary;
}

export async function fetchSources() {
  const response = await fetch(`${apiBase}/v1/sources`, { cache: "no-store" });
  if (!response.ok) throw new Error("failed to fetch sources");
  const payload = (await response.json()) as { sources: SourceSummary[] };
  return payload.sources;
}

export async function fetchAPIKeys() {
  const response = await fetch(`${apiBase}/v1/api-keys`, { cache: "no-store" });
  if (!response.ok) throw new Error("failed to fetch api keys");
  const payload = (await response.json()) as { api_keys: APIKey[] };
  return payload.api_keys;
}

export async function createAPIKey(name: string) {
  const response = await fetch(`${apiBase}/v1/api-keys`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name })
  });
  if (!response.ok) throw new Error("failed to create api key");
  return (await response.json()) as APIKey;
}

export async function revokeAPIKey(id: string) {
  const response = await fetch(`${apiBase}/v1/api-keys/${id}`, {
    method: "DELETE"
  });
  if (!response.ok) throw new Error("failed to revoke api key");
}
