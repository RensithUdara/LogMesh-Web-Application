"use client";

import {
  Activity,
  AlertTriangle,
  BarChart3,
  Bell,
  CheckCircle2,
  ChevronRight,
  Circle,
  ClipboardList,
  Clock3,
  Database,
  FileDown,
  FileText,
  Filter,
  Gauge,
  KeyRound,
  Layers3,
  LogIn,
  LineChart as LineChartIcon,
  Loader2,
  LockKeyhole,
  Pause,
  Play,
  Plus,
  RefreshCw,
  Search,
  Server,
  Settings,
  ShieldCheck,
  TerminalSquare,
  Trash2,
  X
} from "lucide-react";
import {
  FormEvent,
  MouseEvent,
  ReactNode,
  useEffect,
  useMemo,
  useRef,
  useState
} from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis
} from "recharts";
import { authHeaders, bulkIngestLogs, createAPIKey, fetchAPIKeys, fetchAnalytics, fetchRuntime, fetchSources, logExportURL, loginUser, parseTextLog, registerUser, revokeAPIKey } from "@/lib/api";
import { apiBase, environments, levelClass, levels, services } from "@/lib/constants";
import type { APIKey, AnalyticsSummary, CountBucket, LogEvent, LogLevel, RuntimeStats, SearchResponse, SourceSummary, TimelineBucket } from "@/lib/types";

type Tab = "logs" | "analytics" | "sources" | "keys" | "settings";

type Toast = {
  type: "success" | "error";
  message: string;
};

type DashboardStats = {
  total: number;
  errors: number;
  errorRate: string;
  serviceCount: number;
  levelCounts: CountBucket[];
  topServices: CountBucket[];
  topErrors: CountBucket[];
  timeline: TimelineBucket[];
};

const starterLogs = [
  {
    service: "payment-service",
    environment: "production",
    level: "ERROR",
    message: "Payment processing failed",
    host: "server-03",
    trace_id: "pay-7392",
    metadata: { gateway: "stripe", retry: 2, access_token: "demo-token" }
  },
  {
    service: "order-service",
    environment: "production",
    level: "WARN",
    message: "Slow database query",
    host: "server-02",
    trace_id: "ord-1881",
    metadata: { query_time: 1.82, database: "orders" }
  },
  {
    service: "auth-service",
    environment: "staging",
    level: "INFO",
    message: "User session refreshed",
    host: "server-01",
    trace_id: "auth-4410",
    metadata: { provider: "password" }
  }
] satisfies Array<Omit<LogEvent, "id" | "timestamp" | "received_at">>;

export default function Dashboard() {
  const [activeTab, setActiveTab] = useState<Tab>("logs");
  const [logs, setLogs] = useState<LogEvent[]>([]);
  const [selectedLog, setSelectedLog] = useState<LogEvent | null>(null);
  const [query, setQuery] = useState("");
  const [level, setLevel] = useState<"ALL" | LogLevel>("ALL");
  const [service, setService] = useState("ALL");
  const [environment, setEnvironment] = useState("ALL");
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState<Toast | null>(null);
  const [apiStatus, setApiStatus] = useState<"checking" | "online" | "offline">("checking");
  const [analytics, setAnalytics] = useState<AnalyticsSummary | null>(null);
  const [sourceRows, setSourceRows] = useState<SourceSummary[]>([]);
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [runtime, setRuntime] = useState<RuntimeStats | null>(null);
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  const [newKeyName, setNewKeyName] = useState("Production collector");
  const [rawLine, setRawLine] = useState("2026-08-30 10:21:22 ERROR Payment failed while charging card");
  const [authMode, setAuthMode] = useState<"login" | "register">("login");
  const [authForm, setAuthForm] = useState({ email: "demo@logmesh.local", password: "password123" });
  const [authToken, setAuthToken] = useState<string | null>(null);
  const [form, setForm] = useState({
    service: "payment-service",
    environment: "production",
    level: "ERROR" as LogLevel,
    message: "Payment processing failed",
    host: "server-03",
    trace_id: "pay-7392",
    metadata: "{\n  \"gateway\": \"stripe\",\n  \"retry\": 2\n}"
  });
  const searchDebounce = useRef<number | null>(null);

  const fetchLogs = async () => {
    const params = new URLSearchParams({ limit: "500" });
    if (query.trim()) params.set("search", query.trim());
    if (level !== "ALL") params.set("level", level);
    if (service !== "ALL") params.set("service", service);
    if (environment !== "ALL") params.set("environment", environment);

    try {
      const response = await fetch(`${apiBase}/v1/logs?${params.toString()}`, {
        cache: "no-store",
        headers: authHeaders()
      });
      if (!response.ok) throw new Error("search failed");
      const payload = (await response.json()) as SearchResponse;
      setLogs(payload.logs);
      setSelectedLog((current) => payload.logs.find((log) => log.id === current?.id) ?? payload.logs[0] ?? null);
      setApiStatus("online");
      void fetchServerPanels();
    } catch {
      setApiStatus("offline");
      setToast({ type: "error", message: "API unavailable" });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setAuthToken(window.localStorage.getItem("logmesh_token"));
    void bootstrapDashboard();
  }, []);

  useEffect(() => {
    if (apiStatus !== "online") return;
    if (searchDebounce.current) {
      window.clearTimeout(searchDebounce.current);
    }
    searchDebounce.current = window.setTimeout(() => void fetchLogs(), 250);
  }, [apiStatus, query, level, service, environment]);

  useEffect(() => {
    if (apiStatus !== "offline") return;
    const interval = window.setInterval(() => void checkHealth(), 5000);
    return () => window.clearInterval(interval);
  }, [apiStatus]);

  useEffect(() => {
    if (apiStatus !== "online") return;
    if (!autoRefresh) return;
    const interval = window.setInterval(() => void fetchLogs(), 4000);
    return () => window.clearInterval(interval);
  }, [apiStatus, autoRefresh, query, level, service, environment]);

  useEffect(() => {
    if (apiStatus !== "online") return;
    if (!autoRefresh) return;

    const stream = new EventSource(`${apiBase}/v1/stream/logs`);
    const onLog = (event: MessageEvent<string>) => {
      const log = JSON.parse(event.data) as LogEvent;
      if (!matchesActiveFilters(log, { query, level, service, environment })) {
        return;
      }

      setLogs((current) => [log, ...current.filter((item) => item.id !== log.id)].slice(0, 500));
      setSelectedLog((current) => current ?? log);
      void fetchServerPanels();
    };

    stream.addEventListener("log", onLog);
    stream.onerror = () => {
      setApiStatus("offline");
    };

    return () => {
      stream.removeEventListener("log", onLog);
      stream.close();
    };
  }, [apiStatus, autoRefresh, query, level, service, environment]);

  useEffect(() => {
    if (!toast) return;
    const timeout = window.setTimeout(() => setToast(null), 2600);
    return () => window.clearTimeout(timeout);
  }, [toast]);

  const stats = useMemo(() => buildStats(logs), [logs]);

  const submitLog = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    let metadata: Record<string, unknown> | undefined;
    if (form.metadata.trim()) {
      try {
        metadata = JSON.parse(form.metadata) as Record<string, unknown>;
      } catch {
        setToast({ type: "error", message: "Metadata JSON is invalid" });
        return;
      }
    }

    try {
      const response = await fetch(`${apiBase}/v1/logs`, {
        method: "POST",
        headers: { ...authHeaders(), "Content-Type": "application/json" },
        body: JSON.stringify({
          service: form.service,
          environment: form.environment,
          level: form.level,
          message: form.message,
          host: form.host,
          trace_id: form.trace_id,
          metadata
        })
      });

      if (!response.ok) {
        const payload = (await response.json()) as { error?: string };
        throw new Error(payload.error ?? "ingest failed");
      }

      const created = (await response.json()) as LogEvent;
      setToast({ type: "success", message: "Log accepted" });
      setSelectedLog(created);
      await fetchLogs();
      await fetchServerPanels();
    } catch (error) {
      setToast({
        type: "error",
        message: error instanceof Error ? error.message : "Unable to ingest log"
      });
    }
  };

  const seedDemoLogs = async (event: MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    await bulkIngestLogs(starterLogs);
    setToast({ type: "success", message: "Demo logs added" });
    await fetchLogs();
    await fetchServerPanels();
  };

  const submitRawLog = async (event: MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    try {
      const created = await parseTextLog({
        service: form.service,
        environment: form.environment,
        host: form.host,
        trace_id: form.trace_id,
        line: rawLine
      });
      setToast({ type: "success", message: "Raw log parsed" });
      setSelectedLog(created);
      await fetchLogs();
      await fetchServerPanels();
    } catch {
      setToast({ type: "error", message: "Unable to parse raw log" });
    }
  };

  const createKey = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      const key = await createAPIKey(newKeyName);
      setCreatedKey(key.plaintext_key ?? null);
      setToast({ type: "success", message: "API key created" });
      await fetchServerPanels();
    } catch {
      setToast({ type: "error", message: "Unable to create API key" });
    }
  };

  const revokeKey = async (id: string) => {
    try {
      await revokeAPIKey(id);
      setToast({ type: "success", message: "API key revoked" });
      await fetchServerPanels();
    } catch {
      setToast({ type: "error", message: "Unable to revoke API key" });
    }
  };

  const submitAuth = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      const response =
        authMode === "login"
          ? await loginUser(authForm.email, authForm.password)
          : await registerUser(authForm.email, authForm.password);
      window.localStorage.setItem("logmesh_token", response.token);
      setAuthToken(response.token);
      setToast({ type: "success", message: authMode === "login" ? "Signed in" : "Account created" });
    } catch {
      setToast({ type: "error", message: "Auth failed" });
    }
  };

  const logout = () => {
    window.localStorage.removeItem("logmesh_token");
    setAuthToken(null);
  };

  async function bootstrapDashboard() {
    const online = await checkHealth();
    if (!online) {
      setLoading(false);
      return;
    }
    await Promise.all([fetchLogs(), fetchServerPanels()]);
  }

  async function checkHealth() {
    try {
      const response = await fetch(`${apiBase}/healthz`, { cache: "no-store" });
      setApiStatus(response.ok ? "online" : "offline");
      return response.ok;
    } catch {
      setApiStatus("offline");
      return false;
    }
  }

  async function fetchServerPanels() {
    try {
      const [summary, sources, keys, runtimeStats] = await Promise.all([fetchAnalytics(), fetchSources(), fetchAPIKeys(), fetchRuntime()]);
      setAnalytics(summary);
      setSourceRows(sources);
      setApiKeys(keys);
      setRuntime(runtimeStats);
    } catch {
      setApiStatus("offline");
    }
  }

  return (
    <main className="shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">
            <TerminalSquare size={23} />
          </div>
          <div>
            <strong>LogMesh</strong>
            <span>Monitoring Platform</span>
          </div>
        </div>

        <nav className="nav">
          <NavButton active={activeTab === "logs"} icon={<Search size={18} />} label="Logs" onClick={() => setActiveTab("logs")} />
          <NavButton active={activeTab === "analytics"} icon={<BarChart3 size={18} />} label="Analytics" onClick={() => setActiveTab("analytics")} />
          <NavButton active={activeTab === "sources"} icon={<Layers3 size={18} />} label="Sources" onClick={() => setActiveTab("sources")} />
          <NavButton active={activeTab === "keys"} icon={<KeyRound size={18} />} label="API Keys" onClick={() => setActiveTab("keys")} />
          <NavButton active={activeTab === "settings"} icon={<Settings size={18} />} label="Settings" onClick={() => setActiveTab("settings")} />
        </nav>

        <div className="sidebar-status">
          <span className={`status-dot ${apiStatus}`} />
          <div>
            <strong>{apiStatus === "online" ? "Collector online" : apiStatus === "offline" ? "Collector offline" : "Checking"}</strong>
            <span>{apiBase}</span>
          </div>
        </div>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div>
            <h1>{activeTabTitle(activeTab)}</h1>
            <p>{new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(new Date())}</p>
          </div>
          <div className="topbar-actions">
            <button className="icon-button" type="button" title="Refresh logs" onClick={() => void fetchLogs()}>
              <RefreshCw size={18} />
            </button>
            <a className="icon-button" href={logExportURL()} title="Export CSV">
              <FileDown size={18} />
            </a>
            <button className={`toggle ${autoRefresh ? "on" : ""}`} type="button" onClick={() => setAutoRefresh((value) => !value)}>
              {autoRefresh ? <Pause size={16} /> : <Play size={16} />}
              <span>{autoRefresh ? "Live" : "Paused"}</span>
            </button>
            <button className="icon-button" type="button" title="Notifications">
              <Bell size={18} />
            </button>
            <button className="toggle" type="button" onClick={logout} title={authToken ? "Sign out" : "Not signed in"}>
              <LogIn size={16} />
              <span>{authToken ? "Signed In" : "Guest"}</span>
            </button>
          </div>
        </header>

        {!authToken && (
          <form className="panel auth-panel" onSubmit={submitAuth}>
            <div className="panel-title">
              <h2>{authMode === "login" ? "Login" : "Register"}</h2>
              <button className="secondary-button" type="button" onClick={() => setAuthMode(authMode === "login" ? "register" : "login")}>
                {authMode === "login" ? "Register" : "Login"}
              </button>
            </div>
            <div className="form-grid">
              <label>
                Email
                <input
                  autoComplete="username"
                  name="email"
                  type="email"
                  value={authForm.email}
                  onChange={(event) => setAuthForm({ ...authForm, email: event.target.value })}
                />
              </label>
              <label>
                Password
                <input
                  autoComplete={authMode === "login" ? "current-password" : "new-password"}
                  name="password"
                  type="password"
                  value={authForm.password}
                  onChange={(event) => setAuthForm({ ...authForm, password: event.target.value })}
                />
              </label>
            </div>
            <button className="primary-button" type="submit">
              <LogIn size={17} />
              {authMode === "login" ? "Login" : "Create Account"}
            </button>
          </form>
        )}

        <section className="metrics-grid">
          <Metric icon={<Database size={20} />} label="Total Logs" value={(analytics?.total ?? stats.total).toLocaleString()} tone="blue" />
          <Metric icon={<AlertTriangle size={20} />} label="Errors" value={(analytics?.errors ?? stats.errors).toLocaleString()} tone="red" />
          <Metric icon={<Gauge size={20} />} label="Error Rate" value={`${analytics ? analytics.error_rate.toFixed(1) : stats.errorRate}%`} tone="amber" />
          <Metric icon={<Server size={20} />} label="Services" value={(analytics?.service_count ?? stats.serviceCount).toString()} tone="green" />
        </section>

        {activeTab === "logs" && (
          <section className="logs-layout">
            <div className="panel log-panel">
              <div className="toolbar">
                <div className="searchbox">
                  <Search size={18} />
                  <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search messages" />
                </div>
                <Select value={service} onChange={setService} options={["ALL", ...services]} label="Service" />
                <Select value={environment} onChange={setEnvironment} options={["ALL", ...environments]} label="Environment" />
                <Select value={level} onChange={(value) => setLevel(value as "ALL" | LogLevel)} options={levels} label="Level" />
              </div>

              <div className="table-head">
                <span>Time</span>
                <span>Level</span>
                <span>Service</span>
                <span>Message</span>
                <span>Trace</span>
              </div>

              <div className="log-list">
                {loading ? (
                  <div className="empty-state">
                    <Loader2 className="spin" size={24} />
                  </div>
                ) : logs.length === 0 ? (
                  <div className="empty-state">
                    <ClipboardList size={30} />
                    <strong>No logs matched</strong>
                  </div>
                ) : (
                  logs.map((log) => (
                    <button
                      className={`log-row ${selectedLog?.id === log.id ? "selected" : ""}`}
                      key={log.id}
                      type="button"
                      onClick={() => setSelectedLog(log)}
                    >
                      <span>{formatTime(log.timestamp)}</span>
                      <span className={levelClass[log.level]}>{log.level}</span>
                      <span>{log.service}</span>
                      <span>{log.message}</span>
                      <span>{log.trace_id || "-"}</span>
                    </button>
                  ))
                )}
              </div>
            </div>

            <aside className="panel detail-panel">
              <div className="panel-title">
                <h2>Log Detail</h2>
                {selectedLog && (
                  <button className="icon-button small" type="button" title="Clear selection" onClick={() => setSelectedLog(null)}>
                    <X size={16} />
                  </button>
                )}
              </div>

              {selectedLog ? <LogDetail log={selectedLog} /> : <EmptyDetail />}
            </aside>

            <form className="panel ingest-panel" onSubmit={submitLog}>
              <div className="panel-title">
                <h2>Ingest</h2>
                <button type="button" className="secondary-button" onClick={seedDemoLogs}>
                  <Plus size={16} />
                  Demo
                </button>
              </div>

              <div className="form-grid">
                <label>
                  Service
                  <input value={form.service} onChange={(event) => setForm({ ...form, service: event.target.value })} />
                </label>
                <label>
                  Environment
                  <select value={form.environment} onChange={(event) => setForm({ ...form, environment: event.target.value })}>
                    {environments.map((item) => (
                      <option key={item}>{item}</option>
                    ))}
                  </select>
                </label>
                <label>
                  Level
                  <select value={form.level} onChange={(event) => setForm({ ...form, level: event.target.value as LogLevel })}>
                    {levels.slice(1).map((item) => (
                      <option key={item}>{item}</option>
                    ))}
                  </select>
                </label>
                <label>
                  Host
                  <input value={form.host} onChange={(event) => setForm({ ...form, host: event.target.value })} />
                </label>
              </div>

              <label>
                Message
                <input value={form.message} onChange={(event) => setForm({ ...form, message: event.target.value })} />
              </label>

              <label>
                Trace ID
                <input value={form.trace_id} onChange={(event) => setForm({ ...form, trace_id: event.target.value })} />
              </label>

              <label>
                Metadata
                <textarea value={form.metadata} onChange={(event) => setForm({ ...form, metadata: event.target.value })} rows={5} />
              </label>

              <button className="primary-button" type="submit">
                <Plus size={17} />
                Send Log
              </button>

              <div className="raw-log-box">
                <div className="panel-title compact-title">
                  <h2>Raw Text</h2>
                  <FileText size={17} />
                </div>
                <textarea value={rawLine} onChange={(event) => setRawLine(event.target.value)} rows={3} />
                <button className="secondary-button full-width" type="button" onClick={submitRawLog}>
                  <FileText size={16} />
                  Parse Text Log
                </button>
              </div>
            </form>
          </section>
        )}

        {activeTab === "analytics" && <Analytics stats={analytics ? serverStatsToClientStats(analytics) : stats} />}
        {activeTab === "sources" && <Sources sources={sourceRows} />}
        {activeTab === "keys" && (
          <ApiKeys
            apiKeys={apiKeys}
            createdKey={createdKey}
            newKeyName={newKeyName}
            onCreate={createKey}
            onNameChange={setNewKeyName}
            onRevoke={revokeKey}
          />
        )}
        {activeTab === "settings" && <SettingsView runtime={runtime} />}
      </section>

      {toast && (
        <div className={`toast ${toast.type}`}>
          {toast.type === "success" ? <CheckCircle2 size={18} /> : <AlertTriangle size={18} />}
          {toast.message}
        </div>
      )}
    </main>
  );
}

function NavButton({
  active,
  icon,
  label,
  onClick
}: {
  active: boolean;
  icon: ReactNode;
  label: string;
  onClick: () => void;
}) {
  return (
    <button className={`nav-button ${active ? "active" : ""}`} type="button" onClick={onClick}>
      {icon}
      <span>{label}</span>
      {active && <ChevronRight size={17} />}
    </button>
  );
}

function Metric({ icon, label, value, tone }: { icon: ReactNode; label: string; value: string; tone: string }) {
  return (
    <div className={`metric ${tone}`}>
      <div className="metric-icon">{icon}</div>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function Select({
  label,
  options,
  value,
  onChange
}: {
  label: string;
  options: string[];
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <label className="compact-select">
      <span>{label}</span>
      <select value={value} onChange={(event) => onChange(event.target.value)}>
        {options.map((option) => (
          <option key={option}>{option}</option>
        ))}
      </select>
    </label>
  );
}

function LogDetail({ log }: { log: LogEvent }) {
  return (
    <div className="detail">
      <span className={levelClass[log.level]}>{log.level}</span>
      <h3>{log.message}</h3>
      <dl>
        <div>
          <dt>Service</dt>
          <dd>{log.service}</dd>
        </div>
        <div>
          <dt>Environment</dt>
          <dd>{log.environment}</dd>
        </div>
        <div>
          <dt>Host</dt>
          <dd>{log.host || "-"}</dd>
        </div>
        <div>
          <dt>Trace ID</dt>
          <dd>{log.trace_id || "-"}</dd>
        </div>
        <div>
          <dt>Timestamp</dt>
          <dd>{new Date(log.timestamp).toLocaleString()}</dd>
        </div>
      </dl>
      <pre>{JSON.stringify(log.metadata ?? {}, null, 2)}</pre>
    </div>
  );
}

function EmptyDetail() {
  return (
    <div className="empty-detail">
      <Filter size={28} />
      <strong>Select a log</strong>
    </div>
  );
}

function Analytics({ stats }: { stats: DashboardStats }) {
  return (
    <section className="analytics-grid">
      <div className="panel chart-panel wide">
        <div className="panel-title">
          <h2>Logs Over Time</h2>
          <LineChartIcon size={18} />
        </div>
        <ResponsiveContainer width="100%" height={280}>
          <AreaChart data={stats.timeline}>
            <CartesianGrid stroke="#e5e7eb" vertical={false} />
            <XAxis dataKey="time" tickLine={false} axisLine={false} />
            <YAxis allowDecimals={false} tickLine={false} axisLine={false} />
            <Tooltip />
            <Area type="monotone" dataKey="logs" stroke="#2563eb" fill="#bfdbfe" strokeWidth={2} />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <div className="panel chart-panel">
        <div className="panel-title">
          <h2>Levels</h2>
          <Activity size={18} />
        </div>
        <ResponsiveContainer width="100%" height={280}>
          <BarChart data={stats.levelCounts}>
            <CartesianGrid stroke="#e5e7eb" vertical={false} />
            <XAxis dataKey="name" tickLine={false} axisLine={false} />
            <YAxis allowDecimals={false} tickLine={false} axisLine={false} />
            <Tooltip />
            <Bar dataKey="value" fill="#0f766e" radius={[6, 6, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="panel list-panel">
        <div className="panel-title">
          <h2>Top Services</h2>
          <Server size={18} />
        </div>
        {stats.topServices.map((item) => (
          <div className="rank-row" key={item.name}>
            <span>{item.name}</span>
            <strong>{item.value}</strong>
          </div>
        ))}
      </div>

      <div className="panel list-panel">
        <div className="panel-title">
          <h2>Common Errors</h2>
          <AlertTriangle size={18} />
        </div>
        {stats.topErrors.length ? (
          stats.topErrors.map((item) => (
            <div className="rank-row" key={item.name}>
              <span>{item.name}</span>
              <strong>{item.value}</strong>
            </div>
          ))
        ) : (
          <div className="empty-state compact">No errors</div>
        )}
      </div>
    </section>
  );
}

function Sources({ sources }: { sources: SourceSummary[] }) {
  return (
    <section className="panel source-panel">
      <div className="panel-title">
        <h2>Log Sources</h2>
        <Server size={18} />
      </div>
      {sources.map((source) => (
        <div className="source-row" key={`${source.service}-${source.environment}`}>
          <div>
            <strong>{source.service}</strong>
            <span>{source.environment} - {source.host_count} hosts - last seen {new Date(source.last_seen).toLocaleString()}</span>
          </div>
          <strong>{source.log_count.toLocaleString()}</strong>
        </div>
      ))}
      {!sources.length && <div className="empty-state compact">No sources</div>}
    </section>
  );
}

function ApiKeys({
  apiKeys,
  createdKey,
  newKeyName,
  onCreate,
  onNameChange,
  onRevoke
}: {
  apiKeys: APIKey[];
  createdKey: string | null;
  newKeyName: string;
  onCreate: (event: FormEvent<HTMLFormElement>) => void;
  onNameChange: (value: string) => void;
  onRevoke: (id: string) => void;
}) {
  return (
    <section className="panel keys-panel">
      <div className="panel-title">
        <h2>API Keys</h2>
        <KeyRound size={18} />
      </div>

      <form className="key-create" onSubmit={onCreate}>
        <input value={newKeyName} onChange={(event) => onNameChange(event.target.value)} placeholder="Key name" />
        <button className="secondary-button" type="submit">
          <Plus size={16} />
          Create
        </button>
      </form>

      {createdKey && (
        <div className="created-key">
          <strong>New key</strong>
          <code>{createdKey}</code>
        </div>
      )}

      {apiKeys.map((key) => (
        <div className="key-row" key={key.id}>
          <LockKeyhole size={18} />
          <div>
            <strong>{key.name}</strong>
            <span>{key.prefix}************************</span>
          </div>
          <span className={`pill ${key.revoked_at ? "revoked" : "active"}`}>{key.revoked_at ? "Revoked" : "Active"}</span>
          <button className="icon-button small" type="button" title="Revoke key" onClick={() => onRevoke(key.id)} disabled={Boolean(key.revoked_at)}>
            <Trash2 size={16} />
          </button>
        </div>
      ))}
      {!apiKeys.length && <div className="empty-state compact">No API keys</div>}
    </section>
  );
}

function SettingsView({ runtime }: { runtime: RuntimeStats | null }) {
  return (
    <section className="settings-grid">
      <div className="panel settings-card">
        <ShieldCheck size={22} />
        <div>
          <h2>Sensitive Fields</h2>
          <p>password, token, secret, authorization, credit_card, api_key</p>
        </div>
      </div>
      <div className="panel settings-card">
        <Clock3 size={22} />
        <div>
          <h2>Retention</h2>
          <p>Development storage keeps the latest 10,000 logs in memory.</p>
        </div>
      </div>
      <div className="panel settings-card">
        <Circle size={22} />
        <div>
          <h2>Refresh</h2>
          <p>Live mode polls every 4 seconds.</p>
        </div>
      </div>
      <div className="panel settings-card">
        <Gauge size={22} />
        <div>
          <h2>Runtime</h2>
          <p>
            {runtime
              ? `${runtime.go_version}, ${runtime.goroutines} goroutines, ${formatBytes(runtime.heap_alloc_bytes)} heap, ${runtime.stored_logs} logs`
              : "Runtime metrics unavailable."}
          </p>
        </div>
      </div>
    </section>
  );
}

function buildStats(logs: LogEvent[]) {
  const levelMap = new Map<string, number>();
  const serviceMap = new Map<string, number>();
  const errorMap = new Map<string, number>();
  const timelineMap = new Map<string, number>();

  for (const log of logs) {
    levelMap.set(log.level, (levelMap.get(log.level) ?? 0) + 1);
    serviceMap.set(log.service, (serviceMap.get(log.service) ?? 0) + 1);
    if (log.level === "ERROR" || log.level === "FATAL") {
      errorMap.set(log.message, (errorMap.get(log.message) ?? 0) + 1);
    }

    const bucket = new Date(log.timestamp);
    bucket.setSeconds(0, 0);
    const label = bucket.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    timelineMap.set(label, (timelineMap.get(label) ?? 0) + 1);
  }

  const errors = logs.filter((log) => log.level === "ERROR" || log.level === "FATAL").length;

  return {
    total: logs.length,
    errors,
    errorRate: logs.length ? ((errors / logs.length) * 100).toFixed(1) : "0.0",
    serviceCount: serviceMap.size,
    levelCounts: levels
      .filter((item): item is LogLevel => item !== "ALL")
      .map((name) => ({ name, value: levelMap.get(name) ?? 0 })),
    topServices: sortRank(serviceMap),
    topErrors: sortRank(errorMap),
    timeline: Array.from(timelineMap.entries()).map(([time, value]) => ({
      time,
      logs: value
    }))
  };
}

function serverStatsToClientStats(summary: AnalyticsSummary) {
  return {
    total: summary.total,
    errors: summary.errors,
    errorRate: summary.error_rate.toFixed(1),
    serviceCount: summary.service_count,
    levelCounts: summary.level_counts,
    topServices: summary.top_services,
    topErrors: summary.top_errors,
    timeline: summary.timeline
  };
}

function sortRank(map: Map<string, number>) {
  return Array.from(map.entries())
    .map(([name, value]) => ({ name, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 6);
}

function activeTabTitle(tab: Tab) {
  switch (tab) {
    case "analytics":
      return "Analytics";
    case "sources":
      return "Sources";
    case "keys":
      return "API Keys";
    case "settings":
      return "Settings";
    default:
      return "Logs";
  }
}

function formatTime(value: string) {
  return new Date(value).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit"
  });
}

function matchesActiveFilters(
  log: LogEvent,
  filters: {
    query: string;
    level: "ALL" | LogLevel;
    service: string;
    environment: string;
  }
) {
  if (filters.level !== "ALL" && log.level !== filters.level) return false;
  if (filters.service !== "ALL" && log.service !== filters.service) return false;
  if (filters.environment !== "ALL" && log.environment !== filters.environment) return false;
  if (filters.query.trim() && !log.message.toLowerCase().includes(filters.query.trim().toLowerCase())) return false;
  return true;
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${(value / 1024 / 1024).toFixed(1)} MB`;
}
