import type { LogLevel } from "./types";

export const apiBase = process.env.NEXT_PUBLIC_LOGMESH_API_URL ?? "http://localhost:8080";

export const levels: Array<"ALL" | LogLevel> = ["ALL", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"];

export const services = ["payment-service", "order-service", "auth-service", "api-gateway", "worker"];

export const environments = ["production", "staging", "development"];

export const levelClass: Record<LogLevel, string> = {
  TRACE: "level trace",
  DEBUG: "level debug",
  INFO: "level info",
  WARN: "level warn",
  ERROR: "level error",
  FATAL: "level fatal"
};
