import { useDebouncedValue } from "../hooks/useDebouncedValue";
import { useState } from "react";
import { useQuery } from "../hooks/useQuery";
import { QueryState } from "../components/Status";
type Entry = {
  id: number;
  actor: string;
  action: string;
  target: string;
  createdAt: string;
};
const actions: Record<string, string> = {
  setup: "初始化工作台",
  login: "登录",
  password_changed: "修改密码",
  password_reset: "重置密码",
  create_user: "创建成员",
  update_user: "调整成员",
  review: "复核记录",
  bulk_review: "批量复核",
  delete_finding: "删除记录",
  update_settings: "保存配置",
  export: "导出报告",
};
export default function Audit() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [action, setAction] = useState("");
  const debounced = useDebouncedValue(search);
  const q = useQuery<{ data: Entry[]; total: number }>(
    "/audit?" +
      new URLSearchParams({ page: String(page), search: debounced, action }),
  );
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">ACTIVITY LOG</span>
          <h1>审计日志</h1>
          <p>追踪登录、配置修改与证据复核操作。</p>
        </div>
        <button onClick={q.reload}>刷新</button>
      </div>
      <div className="panel">
        <div className="toolbar">
          <input
            aria-label="搜索审计记录"
            placeholder="搜索操作者或操作对象…"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
          <select
            aria-label="审计操作类型"
            value={action}
            onChange={(e) => {
              setAction(e.target.value);
              setPage(1);
            }}
          >
            <option value="">全部操作</option>
            {Object.entries(actions).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
          {(search || action) && (
            <button
              onClick={() => {
                setSearch("");
                setAction("");
                setPage(1);
              }}
            >
              清除筛选
            </button>
          )}
        </div>
        <QueryState {...q} empty={q.data?.data.length === 0} retry={q.reload} />
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>时间</th>
                <th>操作者</th>
                <th>操作</th>
                <th>对象</th>
              </tr>
            </thead>
            <tbody>
              {q.data?.data.map((v) => (
                <tr key={v.id}>
                  <td>{new Date(v.createdAt).toLocaleString()}</td>
                  <td>{v.actor}</td>
                  <td>
                    {actions[v.action] || v.action}
                    <small className="audit-code">{v.action}</small>
                  </td>
                  <td>{v.target}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="pagination">
          <span>共 {q.data?.total || 0} 条</span>
          <button disabled={page === 1} onClick={() => setPage((v) => v - 1)}>
            上一页
          </button>
          <span>{page}</span>
          <button
            disabled={page * 50 >= (q.data?.total || 0)}
            onClick={() => setPage((v) => v + 1)}
          >
            下一页
          </button>
        </div>
      </div>
    </>
  );
}
