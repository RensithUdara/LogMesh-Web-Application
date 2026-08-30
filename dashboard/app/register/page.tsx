"use client";

import { Activity, TerminalSquare, UserPlus } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import { registerUser } from "@/lib/api";

type RegisterState = {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
};

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState<RegisterState>({
    name: "",
    email: "",
    password: "",
    confirmPassword: ""
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

    if (form.password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }
    if (form.password !== form.confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    setLoading(true);
    try {
      const response = await registerUser(form.name, form.email, form.password);
      window.localStorage.setItem("logmesh_token", response.token);
      router.replace("/dashboard");
    } catch {
      setError("Unable to create account.");
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
          <h1>Create Account</h1>
          <p>Set up your first project workspace.</p>
        </div>

        <form className="auth-form" onSubmit={submit}>
          <label>
            Name
            <input
              autoComplete="name"
              name="name"
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
              required
            />
          </label>

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
              autoComplete="new-password"
              name="password"
              type="password"
              value={form.password}
              onChange={(event) => setForm({ ...form, password: event.target.value })}
              required
            />
          </label>

          <label>
            Confirm Password
            <input
              autoComplete="new-password"
              name="confirmPassword"
              type="password"
              value={form.confirmPassword}
              onChange={(event) => setForm({ ...form, confirmPassword: event.target.value })}
              required
            />
          </label>

          {error && <p className="auth-error">{error}</p>}

          <button className="primary-button" type="submit" disabled={loading}>
            {loading ? <Activity className="spin" size={17} /> : <UserPlus size={17} />}
            Register
          </button>
        </form>

        <p className="auth-switch">
          Already have an account? <Link href="/">Login</Link>
        </p>
      </section>
    </main>
  );
}
