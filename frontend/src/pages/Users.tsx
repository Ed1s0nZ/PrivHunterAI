import Modal from "../components/Modal";
import { useState } from "react";
import { api, send, type User } from "../api/client";
import { useQuery } from "../hooks/useQuery";
import { QueryState } from "../components/Status";
export default function Users({ currentUser }: { currentUser: User }) {
  const q = useQuery<{ data: User[] }>("/users");
  const [createOpen, setCreateOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [message, setMessage] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("viewer");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">ACCESS CONTROL</span>
          <h1>用户管理</h1>
          <p>最小权限访问，账号调整后已有会话将失效。</p>
        </div>
        <button
          className="primary"
          onClick={() => {
            setError("");
            setCreateOpen(true);
          }}
        >
          ＋ 添加成员
        </button>
      </div>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {message && (
        <p className="success" role="status">
          {message}
        </p>
      )}
      <div className="member-summary">
        <div>
          <span>工作台成员</span>
          <b>{q.data?.data.length || 0}</b>
        </div>
        <div>
          <span>管理员</span>
          <b>
            {q.data?.data.filter((v) => v.role === "admin" && !v.disabled)
              .length || 0}
          </b>
        </div>
        <div>
          <span>启用账号</span>
          <b>{q.data?.data.filter((v) => !v.disabled).length || 0}</b>
        </div>
        <p>
          按职责分配权限，
          <br />
          协作过程可在审计日志追踪。
        </p>
      </div>
      <div className="panel member-panel">
        <div className="section-heading">
          <h2>成员列表</h2>
          <span className="count-pill">{q.data?.data.length || 0} 位成员</span>
        </div>
        <div className="toolbar">
          <input
            aria-label="搜索成员"
            placeholder="搜索用户名…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <small>管理员可调整权限；始终保留至少一名启用的管理员。</small>
        </div>
        <QueryState {...q} retry={q.reload} />
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>用户名</th>
                <th>角色</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {q.data?.data
                .filter((u) =>
                  u.username.toLowerCase().includes(query.toLowerCase()),
                )
                .map((u) => (
                  <tr key={u.id}>
                    <td>
                      <div className="member-identity">
                        <span className="member-avatar">
                          {u.username[0].toUpperCase()}
                        </span>
                        <div>
                          <b>{u.username}</b>
                          <small>
                            {u.id === currentUser.id
                              ? "当前账号"
                              : "工作台成员"}
                          </small>
                        </div>
                      </div>
                    </td>
                    <td>
                      <select
                        aria-label={`${u.username} 的角色`}
                        value={u.role}
                        disabled={busy || u.id === currentUser.id}
                        onChange={async (e) => {
                          const next = e.target.value;
                          if (
                            !confirm(
                              `将 ${u.username} 调整为${next === "admin" ? "管理员" : "只读成员"}？已有会话将失效。`,
                            )
                          )
                            return;
                          setBusy(true);
                          setError("");
                          try {
                            await api(
                              "/users/" + u.id,
                              send("PATCH", {
                                role: next,
                                disabled: u.disabled,
                              }),
                            );
                            setMessage("角色已更新");
                            q.reload();
                          } catch (e) {
                            setError((e as Error).message);
                          } finally {
                            setBusy(false);
                          }
                        }}
                      >
                        <option value="admin">管理员</option>
                        <option value="viewer">只读成员</option>
                      </select>
                    </td>
                    <td>
                      <span
                        className={
                          "badge " + (u.disabled ? "" : "badge-resolved")
                        }
                      >
                        {u.disabled ? "已禁用" : "启用中"}
                      </span>
                    </td>
                    <td>
                      <button
                        disabled={busy || u.id === currentUser.id}
                        title={
                          u.id === currentUser.id
                            ? "当前账号不能在这里停用自身"
                            : undefined
                        }
                        onClick={async () => {
                          if (
                            !u.disabled &&
                            !confirm(`停用 ${u.username}？已有会话将失效。`)
                          )
                            return;
                          setBusy(true);
                          setError("");
                          try {
                            await api(
                              "/users/" + u.id,
                              send("PATCH", {
                                role: u.role,
                                disabled: !u.disabled,
                              }),
                            );
                            q.reload();
                          } catch (e) {
                            setError((e as Error).message);
                          } finally {
                            setBusy(false);
                          }
                        }}
                      >
                        {u.disabled ? "启用" : "禁用"}
                      </button>
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      </div>
      <div className="permission-note">
        <span>ⓘ</span>
        <p>
          管理员可管理配置、成员与证据；只读成员可查看和导出检测结果。当前账号与最后一名管理员受到保护。
        </p>
      </div>
      {createOpen && (
        <Modal
          title="添加成员"
          busy={busy}
          onClose={() => {
            setCreateOpen(false);
            setPassword("");
          }}
        >
          <form
            className="form-grid member-create"
            onSubmit={async (e) => {
              e.preventDefault();
              setBusy(true);
              setError("");
              try {
                await api(
                  "/users",
                  send("POST", { username: name, password, role }),
                );
                setMessage("成员已创建，可使用初始密码登录");
                setCreateOpen(false);
                setName("");
                setPassword("");
                q.reload();
              } catch (e) {
                setError((e as Error).message);
              } finally {
                setBusy(false);
              }
            }}
          >
            <div className="section-heading">
              <div>
                <h2>添加成员</h2>
                <small>创建一个工作台账号，并指定访问权限。</small>
              </div>
              <button
                type="button"
                disabled={busy}
                onClick={() => {
                  setCreateOpen(false);
                  setPassword("");
                }}
              >
                关闭
              </button>
            </div>
            <label>
              用户名
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                minLength={3}
              />
            </label>
            <label>
              初始密码
              <input
                type="password"
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={12}
              />
            </label>
            <label>
              角色
              <select value={role} onChange={(e) => setRole(e.target.value)}>
                <option value="viewer">只读成员</option>
                <option value="admin">管理员</option>
              </select>
            </label>
            {error && <p className="error">{error}</p>}
            <button className="primary" disabled={busy}>
              创建用户
            </button>
          </form>
        </Modal>
      )}
    </>
  );
}
