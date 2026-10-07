import { useState } from "react";
import { api, send, type Session } from "../api/client";
export default function Login({
  setup,
  onLogin,
  onSetup,
}: {
  setup: boolean;
  onLogin: (s: Session) => void;
  onSetup: () => void;
}) {
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  return (
    <div className="login">
      <div className="login-surface">
        <div className="login-story">
          <div className="login-brand">
            <b>PH</b>
            <span>
              PrivHunter<small>安全分析工作台</small>
            </span>
          </div>
          <span className="eyebrow">SECURITY OPERATIONS</span>
          <h1>
            看清权限边界。
            <br />
            让证据说话。
          </h1>
          <p>
            从被动流量到权限复核，将每一次检测变成可追踪、可验证的安全记录。
          </p>
          <div className="login-capabilities">
            <span>被动流量检测</span>
            <span>证据对比复核</span>
            <span>本地数据存储</span>
          </div>
        </div>
        <form
          className="login-card"
          onSubmit={async (e) => {
            e.preventDefault();
            setBusy(true);
            setError("");
            try {
              if (setup) {
                await api(
                  "/auth/setup",
                  send("POST", { username: name, password }),
                );
                setPassword("");
                onSetup();
              } else
                onLogin(
                  await api<Session>(
                    "/auth/login",
                    send("POST", { username: name, password }),
                  ),
                );
            } catch (e) {
              setError((e as Error).message);
            } finally {
              setBusy(false);
            }
          }}
        >
          <span className="eyebrow">WORKSPACE ACCESS</span>
          <h2>{setup ? "创建管理员" : "欢迎回来"}</h2>
          <p>
            {setup
              ? "首次启动，请设置工作台管理员。"
              : "登录后访问检测记录与扫描配置。"}
          </p>
          <label>
            用户名
            <input
              autoComplete="username"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              minLength={3}
            />
          </label>
          <label>
            密码
            <input
              type="password"
              autoComplete={setup ? "new-password" : "current-password"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={setup ? 12 : 1}
            />
          </label>
          {setup && <small>密码至少 12 字节，最长 72 字节。</small>}
          {error && (
            <p className="error" role="alert">
              {error}
            </p>
          )}
          <button className="primary" disabled={busy}>
            {busy ? "正在处理…" : setup ? "初始化工作台" : "登录工作台 →"}
          </button>
          <small>会话受保护 · 数据保存在本地</small>
        </form>
      </div>
    </div>
  );
}
