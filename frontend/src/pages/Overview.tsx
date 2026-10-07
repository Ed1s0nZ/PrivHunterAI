import { useEffect } from "react";
import { useQuery } from "../hooks/useQuery";
import { Badge, QueryState } from "../components/Status";
import type { FindingSummary } from "../api/client";
import type { Page } from "../components/Layout";
export type ResultSelection = {
  result?: string;
  review?: string;
  search?: string;
};
type Dashboard = {
  counts: Record<string, number>;
  summary: {
    reviews: Record<string, number>;
    trend: { date: string; total: number; suspected: number }[];
    targets: { host: string; total: number; pending: number }[];
  };
  recent: FindingSummary[];
  readiness?: {
    scope: boolean;
    identity: boolean;
    model: boolean;
    paused: boolean;
  };
  runtime?: { queued: number; processed: number; dropped: number };
};
export default function Overview({
  navigate,
}: {
  navigate: (page: Page, selection?: ResultSelection) => void;
}) {
  const q = useQuery<Dashboard>("/dashboard");
  useEffect(() => {
    const timer = setInterval(() => {
      if (document.visibilityState === "visible") q.reload();
    }, 30000);
    return () => clearInterval(timer);
  }, []);
  const data = q.data;
  const max = Math.max(1, ...(data?.summary.trend.map((v) => v.total) || []));
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">WORKSPACE / OVERVIEW</span>
          <h1>安全态势概览</h1>
          <p>先确认采集准备，再处理待复核证据。数据每 30 秒自动更新。</p>
        </div>
        <div className="toolbar">
          <button onClick={q.reload}>刷新数据</button>
          <button
            className="primary"
            onClick={() => navigate("findings", { review: "pending" })}
          >
            进入复核队列 →
          </button>
        </div>
      </div>
      <QueryState {...q} retry={q.reload} />
      {data && (
        <>
          {data.readiness && (
            <div className="readiness-strip">
              <div>
                <span
                  className={
                    "live-dot " + (!data.readiness.paused ? "running" : "")
                  }
                />
                <b>{data.readiness.paused ? "扫描已暂停" : "扫描已启用"}</b>
                <small>
                  {data.runtime?.queued || 0} 个排队 · 本次处理{" "}
                  {data.runtime?.processed || 0} 条
                </small>
              </div>
              <button onClick={() => navigate("scanner")}>
                {data.readiness.scope && data.readiness.identity
                  ? "管理采集"
                  : "完成采集配置"}{" "}
                →
              </button>
            </div>
          )}
          <div className="stat-grid">
            {[
              {
                label: "检测总数",
                value: data.counts.total,
                sub: "全部已保存证据",
                selection: {},
              },
              {
                label: "待人工复核",
                value: data.summary.reviews.pending,
                sub: "尚未形成复核结论",
                selection: { review: "pending" },
              },
              {
                label: "已确认风险",
                value: data.summary.reviews.confirmed,
                sub: "人工确认后的记录",
                selection: { review: "confirmed" },
              },
              {
                label: "检测失败",
                value: data.counts.error,
                sub: "查看原因，检查配置",
                selection: { result: "error" },
              },
            ].map((item) => (
              <button
                className="stat stat-action"
                key={item.label}
                onClick={() => navigate("findings", item.selection)}
              >
                <span>
                  {item.label}
                  <span>↗</span>
                </span>
                <strong>{item.value || 0}</strong>
                <small>{item.sub}</small>
              </button>
            ))}
          </div>
          <div className="dashboard-grid">
            <section className="panel trend-panel">
              <div className="section-heading">
                <div>
                  <h2>近 7 日检测趋势</h2>
                  <small>按北京时间 · 已保存检测记录</small>
                </div>
                <span className="chart-key">
                  <i />
                  检测总数 <i className="suspect" />
                  疑似越权
                </span>
              </div>
              <div
                className="trend-chart"
                role="img"
                aria-label={data.summary.trend
                  .map((v) => `${v.date} 检测 ${v.total}，疑似 ${v.suspected}`)
                  .join("；")}
              >
                {data.summary.trend.map((v) => (
                  <div className="trend-day" key={v.date}>
                    <span className="trend-count">{v.total}</span>
                    <div className="trend-track">
                      <div
                        className="trend-bar"
                        style={{ height: `${(v.total / max) * 100}%` }}
                      />
                      <div
                        className="trend-bar suspect"
                        style={{ height: `${(v.suspected / max) * 100}%` }}
                      />
                    </div>
                    <small>{v.date.slice(5)}</small>
                  </div>
                ))}
              </div>
              {data.counts.total === 0 && (
                <p className="chart-note">
                  暂无检测数据。配置范围并连接上游代理后，这里将显示真实趋势。
                </p>
              )}
            </section>
            <section className="panel">
              <div className="section-heading">
                <div>
                  <h2>判定与复核进度</h2>
                  <small>疑似结果仍需人工验证</small>
                </div>
              </div>
              <div className="distribution">
                {[
                  ["true", "疑似越权"],
                  ["false", "权限拒绝 / 未发现"],
                  ["unknown", "待分析"],
                  ["error", "失败"],
                ].map(([key, label]) => (
                  <button
                    key={key}
                    onClick={() => navigate("findings", { result: key })}
                  >
                    <span>
                      <i className={"result-dot dot-" + key} />
                      {label}
                    </span>
                    <b>{data.counts[key] || 0}</b>
                  </button>
                ))}
              </div>
              <div className="review-progress">
                <span>
                  已完成复核{" "}
                  <b>
                    {data.counts.total
                      ? (
                          (100 *
                            (data.counts.total -
                              data.summary.reviews.pending)) /
                          data.counts.total
                        ).toFixed(0)
                      : 0}
                    %
                  </b>
                </span>
                <progress
                  max={Math.max(1, data.counts.total)}
                  value={data.counts.total - data.summary.reviews.pending}
                />
                <small>
                  已解决 {data.summary.reviews.resolved} · 误报{" "}
                  {data.summary.reviews.false_positive}
                </small>
              </div>
            </section>
          </div>
          <div className="dashboard-grid">
            <section className="panel">
              <div className="section-heading">
                <div>
                  <h2>最近检测</h2>
                  <small>按检测时间排序</small>
                </div>
                <button
                  className="text-button"
                  onClick={() => navigate("findings")}
                >
                  查看全部 →
                </button>
              </div>
              {data.recent.length ? (
                <div className="recent-list">
                  {data.recent.map((v) => (
                    <button
                      key={v.id}
                      onClick={() => navigate("findings", { search: v.url })}
                    >
                      <span className="method">{v.method}</span>
                      <span className="recent-url">
                        {v.url}
                        <small>{new Date(v.createdAt).toLocaleString()}</small>
                      </span>
                      <Badge value={v.result} />
                    </button>
                  ))}
                </div>
              ) : (
                <div className="action-empty">
                  <span>⌕</span>
                  <h3>第一条证据，从一次采集开始</h3>
                  <p>
                    {data.readiness
                      ? "完成目标范围和账号 B 身份配置，再通过代理访问测试接口。"
                      : "工作台尚无检测记录，管理员完成采集后即可在这里查看证据。"}
                  </p>
                  <button
                    onClick={() =>
                      navigate(data.readiness ? "scanner" : "findings")
                    }
                  >
                    {data.readiness ? "设置采集范围" : "查看检测结果"} →
                  </button>
                </div>
              )}
            </section>
            <section className="panel">
              <div className="section-heading">
                <div>
                  <h2>目标复核热点</h2>
                  <small>按待复核数量排序 · 前 5 个域名</small>
                </div>
              </div>
              {data.summary.targets.length ? (
                <div className="target-list">
                  {data.summary.targets.map((v) => (
                    <button
                      key={v.host}
                      onClick={() =>
                        navigate("findings", {
                          search: v.host,
                          review: "pending",
                        })
                      }
                    >
                      <span>
                        {v.host}
                        <small>累计 {v.total} 条证据</small>
                      </span>
                      <b>
                        {v.pending}
                        <small>待复核</small>
                      </b>
                    </button>
                  ))}
                </div>
              ) : (
                <div className="action-empty compact">
                  <span>◎</span>
                  <h3>尚未采集目标</h3>
                  <p>这里将帮助你识别积压最多的目标。</p>
                </div>
              )}
            </section>
          </div>
          {data.readiness && (
            <section className="panel">
              <div className="section-heading">
                <div>
                  <h2>采集就绪检查</h2>
                  <small>按顺序完成配置，模型分析为可选步骤。</small>
                </div>
              </div>
              <div className="setup-grid">
                {[
                  {
                    title: "目标范围",
                    ready: data.readiness.scope,
                    text: "明确允许检测的域名或 IP",
                    page: "scanner",
                  },
                  {
                    title: "账号 B 身份",
                    ready: data.readiness.identity,
                    text: "设置用于权限对比的请求头",
                    page: "scanner",
                  },
                  {
                    title: "分析模型",
                    ready: data.readiness.model,
                    text: "未配置时保留证据供人工复核",
                    page: "settings",
                  },
                ].map((v, i) => (
                  <button
                    key={v.title}
                    onClick={() => navigate(v.page as Page)}
                  >
                    <span className={"step-marker " + (v.ready ? "done" : "")}>
                      {v.ready ? "✓" : i + 1}
                    </span>
                    <div>
                      <b>{v.title}</b>
                      <small>{v.text}</small>
                    </div>
                    <span>{v.ready ? "已配置" : "去配置"} →</span>
                  </button>
                ))}
              </div>
            </section>
          )}
        </>
      )}
    </>
  );
}
