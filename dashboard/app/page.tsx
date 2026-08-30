"use client";

import { Activity, LogIn, TerminalSquare } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import { loginUser } from "@/lib/api";

type AuthState = {
  email: string;
  password: string;
};

export default function LoginPage() {
  const router = useRouter();
  const [form, setForm] = useState<AuthState>({
    email: "demo@logmesh.local",
    password: "password123"
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (window.localStorage.getItem("logmesh_token")) {
      router.replace("/dashboard");
    }
  }, [router]);

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    setLoading(true);

    try {
      const response = await loginUser(form.email, form.password);
      window.localStorage.setItem("logmesh_token", response.token);
      router.replace("/dashboard");
    } catch {
      setError("Invalid email or password.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="auth-screen">
      <section className="auth-card">
        <div className="auth-brand">
          <div className="brand-mark">
            <TerminalSquare size={24} />
          </div>
          <div>
            <strong>LogMesh</strong>
            <span>Distributed Log Monitoring</span>
          </div>
        </div>

        <div className="auth-heading">
          <h1>Login</h1>
          <p>Open your monitoring workspace.</p>
        </div>

        <form className="auth-form" onSubmit={submit}>
          <label>
            Email
            <input
              autoComplete="username"
              name="email"
              type="email"
              value={form.email}
              onChange={(event) => setForm({ ...form, email: event.target.value })}
              required
            />
          </label>

          <label>
            Password
            <input
              autoComplete="current-password"
              name="password"
              type="password"
              value={form.password}
              onChange={(event) => setForm({ ...form, password: event.target.value })}
              required
            />
          </label>

          {error && <p className="auth-error">{error}</p>}

          <button className="primary-button" type="submit" disabled={loading}>
            {loading ? <Activity className="spin" size={17} /> : <LogIn size={17} />}
            Login
          </button>
        </form>

        <p className="auth-switch">
          New to LogMesh? <Link href="/register">Create an account</Link>
        </p>
      </section>
    </main>
  );
}
