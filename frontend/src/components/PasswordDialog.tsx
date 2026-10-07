import Modal from "./Modal";
import { useState } from "react";
import { api, send } from "../api/client";
export default function PasswordDialog({
  close,
  changed,
}: {
  close: () => void;
  changed: () => void;
}) {
  const [current, setCurrent] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  return (
    <Modal title="修改密码" onClose={close} busy={busy}>
      <form
        className="form-grid"
        onSubmit={async (e) => {
          e.preventDefault();
          setBusy(true);
          setError("");
          try {
            await api("/auth/password", send("POST", { current, password }));
            changed();
          } catch (e) {
            setError((e as Error).message);
          } finally {
            setBusy(false);
          }
        }}
      >
        <div className="page-heading">
          <h2>修改密码</h2>
          <button type="button" disabled={busy} onClick={close}>
            关闭
          </button>
        </div>
        <label>
          当前密码
          <input
            autoFocus
            type="password"
            autoComplete="current-password"
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
            required
          />
        </label>
        <label>
          新密码
          <input
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={12}
          />
        </label>
        <small>修改后所有现有会话失效，请使用新密码登录。</small>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <button className="primary" disabled={busy}>
          保存新密码
        </button>
      </form>
    </Modal>
  );
}
