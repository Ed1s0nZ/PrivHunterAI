import Modal from "../components/Modal";
import { useDebouncedValue } from "../hooks/useDebouncedValue";
import EvidenceComparison from "../components/EvidenceComparison";
import { useEffect, useState } from "react";
import { api, send, type Finding, type FindingSummary } from "../api/client";
import { useQuery } from "../hooks/useQuery";
import { Badge, QueryState, labels } from "../components/Status";
export default function Findings({
  admin,
  initial = {},
}: {
  admin: boolean;
  initial?: { search?: string; result?: string; review?: string };
}) {
  const [search, setSearch] = useState(initial.search || "");
  const [result, setResult] = useState(initial.result || "");
  const [review, setReview] = useState(initial.review || "");
  const [checked, setChecked] = useState<number[]>([]);
  const [bulkStatus, setBulkStatus] = useState("confirmed");
  const [notice, setNotice] = useState("");
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<FindingSummary | null>(null);
  const detailQuery = useQuery<Finding>(
    selected ? "/findings/" + selected.id : null,
  );
  const [note, setNote] = useState("");
  const [status, setStatus] = useState("pending");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const debouncedSearch = useDebouncedValue(search);
  const q = useQuery<{ data: FindingSummary[]; total: number }>(
    "/findings?" +
      new URLSearchParams({
        search: debouncedSearch,
        result,
        review,
        page: String(page),
        size: "20",
      }),
  );
  useEffect(() => {
    setChecked([]);
  }, [search, result, review, page]);
  const bulk = async () => {
    setBusy(true);
    setError("");
    try {
      await api(
        "/findings/bulk-review",
        send("POST", { ids: checked, review: bulkStatus }),
      );
      setNotice(`已更新 ${checked.length} 条记录`);
      setChecked([]);
      q.reload();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  const mutate = async (method: string, body?: unknown) => {
    if (!selected) return;
    setBusy(true);
    setError("");
    try {
      await api("/findings/" + selected.id, send(method, body || {}));
      setSelected(null);
      q.reload();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">EVIDENCE & REVIEW</span>
          <h1>检测结果</h1>
          <p>检索权限风险，对比证据并记录结论。</p>
        </div>
        <div className="toolbar">
          <a
            className="download"
            href={
              "/api/export?" +
              new URLSearchParams({ search, result, review, format: "csv" })
            }
          >
            导出 CSV
          </a>
          <a
            className="download"
            href={
              "/api/export?" +
              new URLSearchParams({ search, result, review, format: "json" })
            }
          >
            导出 JSON
          </a>
          <button onClick={q.reload}>刷新</button>
        </div>
      </div>
      <div className="panel">
        <div className="toolbar">
          <input
            aria-label="搜索 URL"
            placeholder="搜索 URL…"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
          <select
            aria-label="检测判定"
            value={result}
            onChange={(e) => {
              setResult(e.target.value);
              setPage(1);
            }}
          >
            <option value="">全部判定</option>
            {["true", "false", "unknown", "error"].map((v) => (
              <option key={v} value={v}>
                {labels[v]}
              </option>
            ))}
          </select>
          <select
            aria-label="复核状态"
            value={review}
            onChange={(e) => {
              setReview(e.target.value);
              setPage(1);
            }}
          >
            <option value="">全部复核状态</option>
            {["pending", "confirmed", "false_positive", "resolved"].map((v) => (
              <option key={v} value={v}>
                {labels[v]}
              </option>
            ))}
          </select>
        </div>
        {notice && (
          <p className="success" role="status">
            {notice}
          </p>
        )}
        {error && !selected && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        {(search || result || review) && (
          <button
            className="filter-reset"
            onClick={() => {
              setSearch("");
              setResult("");
              setReview("");
              setPage(1);
            }}
          >
            清除筛选条件 ×
          </button>
        )}
        {admin && checked.length > 0 && (
          <div className="bulk-toolbar">
            <span>已选 {checked.length} 条（当前页）</span>
            <select
              aria-label="批量复核状态"
              value={bulkStatus}
              onChange={(e) => setBulkStatus(e.target.value)}
            >
              {["pending", "confirmed", "false_positive", "resolved"].map(
                (v) => (
                  <option key={v} value={v}>
                    {labels[v]}
                  </option>
                ),
              )}
            </select>
            <button className="primary" disabled={busy} onClick={bulk}>
              应用复核状态
            </button>
            <button onClick={() => setChecked([])}>取消选择</button>
          </div>
        )}
        <QueryState {...q} empty={q.data?.data.length === 0} retry={q.reload} />
        {!q.loading && !q.error && !!q.data?.data.length && (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>
                    {admin && (
                      <input
                        type="checkbox"
                        aria-label="选择当前页全部记录"
                        checked={
                          !!q.data?.data.length &&
                          checked.length === q.data.data.length
                        }
                        onChange={(e) =>
                          setChecked(
                            e.target.checked
                              ? q.data!.data.map((v) => v.id)
                              : [],
                          )
                        }
                      />
                    )}{" "}
                    请求
                  </th>
                  <th>判定</th>
                  <th>复核</th>
                  <th>可信度</th>
                  <th>检测时间</th>
                </tr>
              </thead>
              <tbody>
                {q.data.data.map((v) => (
                  <tr
                    key={v.id}
                    onClick={() => {
                      setSelected(v);
                      setNote(v.note);
                      setStatus(v.review);
                      setError("");
                    }}
                  >
                    <td>
                      {admin && (
                        <input
                          type="checkbox"
                          aria-label={`选择记录 ${v.id}`}
                          checked={checked.includes(v.id)}
                          onClick={(e) => e.stopPropagation()}
                          onChange={(e) =>
                            setChecked((prev) =>
                              e.target.checked
                                ? [...prev, v.id]
                                : prev.filter((id) => id !== v.id),
                            )
                          }
                        />
                      )}
                      <button
                        className="text-button"
                        onClick={() => {
                          setSelected(v);
                          setNote(v.note);
                          setStatus(v.review);
                        }}
                      >
                        <span className="method">{v.method}</span> {v.url}
                      </button>
                    </td>
                    <td>
                      <Badge value={v.result} />
                    </td>
                    <td>
                      <Badge value={v.review} />
                    </td>
                    <td>{v.confidence || "—"}</td>
                    <td>{new Date(v.createdAt).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <div className="pagination">
          <span>共 {q.data?.total || 0} 条</span>
          <button disabled={page === 1} onClick={() => setPage((p) => p - 1)}>
            上一页
          </button>
          <span>{page}</span>
          <button
            disabled={page * 20 >= (q.data?.total || 0)}
            onClick={() => setPage((p) => p + 1)}
          >
            下一页
          </button>
        </div>
      </div>
      {selected && (
        <Modal title="检测证据" busy={busy} onClose={() => setSelected(null)}>
          <div className="page-heading">
            <h2>检测证据 #{selected.id}</h2>
            <button autoFocus disabled={busy} onClick={() => setSelected(null)}>
              关闭
            </button>
          </div>
          <p className="url">
            {selected.method} {selected.url}
          </p>
          <Badge value={selected.result} />
          <p>{selected.reason}</p>
          <QueryState {...detailQuery} retry={detailQuery.reload} />
          {detailQuery.data && (
            <EvidenceComparison finding={detailQuery.data} />
          )}
          {admin && (
            <div className="review-form">
              <label>
                复核状态
                <select
                  value={status}
                  onChange={(e) => setStatus(e.target.value)}
                >
                  {["pending", "confirmed", "false_positive", "resolved"].map(
                    (v) => (
                      <option key={v} value={v}>
                        {labels[v]}
                      </option>
                    ),
                  )}
                </select>
              </label>
              <label>
                备注
                <textarea
                  maxLength={4000}
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                />
              </label>
              {error && <p className="error">{error}</p>}
              <button
                className="primary"
                disabled={busy}
                onClick={() => mutate("PATCH", { review: status, note })}
              >
                保存复核
              </button>
              <button
                className="danger"
                disabled={busy}
                onClick={() => {
                  if (confirm("删除这条检测记录？")) mutate("DELETE");
                }}
              >
                删除记录
              </button>
            </div>
          )}
        </Modal>
      )}
    </>
  );
}
