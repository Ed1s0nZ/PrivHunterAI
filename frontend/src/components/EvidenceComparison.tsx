import { useMemo, useState } from "react";
import type { Finding } from "../api/client";
function fields(value: unknown, prefix = "", out: Record<string, string> = {}) {
  if (Object.keys(out).length >= 500) return out;
  if (value && typeof value === "object") {
    for (const [key, v] of Object.entries(value)) {
      fields(v, prefix ? prefix + "." + key : key, out);
    }
  } else out[prefix || "$"] = JSON.stringify(value) ?? "null";
  return out;
}
export default function EvidenceComparison({ finding }: { finding: Finding }) {
  const [tab, setTab] = useState<"response" | "request" | "diff">("response");
  const [message, setMessage] = useState("");
  const diff = useMemo(() => {
    try {
      const a = fields(JSON.parse(finding.responseA));
      const b = fields(JSON.parse(finding.responseB));
      return [...new Set([...Object.keys(a), ...Object.keys(b)])]
        .filter((k) => a[k] !== b[k])
        .map((k) => ({
          path: k,
          a: a[k] ?? "（不存在）",
          b: b[k] ?? "（不存在）",
        }))
        .slice(0, 500);
    } catch {
      return null;
    }
  }, [finding]);
  const values =
    tab === "request"
      ? [finding.requestA || "", finding.requestB || ""]
      : [finding.responseA || "", finding.responseB || ""];
  return (
    <>
      <div className="evidence-tabs" role="tablist" aria-label="证据视图">
        {[
          ["response", "响应对比"],
          ["request", "请求对比"],
          ["diff", "JSON 字段差异"],
        ].map(([key, label]) => (
          <button
            key={key}
            role="tab"
            aria-selected={tab === key}
            onClick={() => setTab(key as typeof tab)}
          >
            {label}
            {key === "diff" && diff !== null ? " · " + diff.length : ""}
          </button>
        ))}
      </div>
      {tab === "diff" ? (
        <div className="field-diff">
          {diff === null ? (
            <p>响应不是可解析的 JSON，请使用响应对比查看原文。</p>
          ) : diff.length === 0 ? (
            <p>
              未发现可见字段差异。脱敏后的内容一致不代表权限判断已得到证实。
            </p>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>字段路径</th>
                    <th>账号 A</th>
                    <th>账号 B</th>
                  </tr>
                </thead>
                <tbody>
                  {diff.map((d) => (
                    <tr key={d.path}>
                      <td>{d.path}</td>
                      <td>
                        <code>{d.a}</code>
                      </td>
                      <td>
                        <code>{d.b}</code>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
          <small>最多展示 500 个差异字段；脱敏字段不会恢复原值。</small>
        </div>
      ) : (
        <div className="comparison">
          {values.map((body, i) => (
            <div key={i}>
              <div className="section-heading">
                <h3>
                  账号 {i === 0 ? "A" : "B"}{" "}
                  {tab === "request" ? "请求" : "响应"}
                </h3>
                <button
                  onClick={async () => {
                    try {
                      await navigator.clipboard.writeText(body);
                      setMessage("已复制脱敏证据");
                    } catch {
                      setMessage("复制失败，请手动选择文本");
                    }
                  }}
                >
                  复制
                </button>
              </div>
              <pre>{body || "无内容"}</pre>
            </div>
          ))}
        </div>
      )}
      {message && <small role="status">{message}</small>}
    </>
  );
}
