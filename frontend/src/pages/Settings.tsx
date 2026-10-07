import { useUnsavedChanges } from "../hooks/useUnsavedChanges";
import { useEffect, useState } from "react";
import { api, send } from "../api/client";
import { useQuery } from "../hooks/useQuery";
import { QueryState } from "../components/Status";
type Config = {
  revision: number;
  paused: boolean;
  domains: string[];
  endpoint: string;
  model: string;
  skipSuffixes: string[];
  denyKeywords: string[];
};
export default function Settings({ scanner = false }: { scanner?: boolean }) {
  const runtime = useQuery<{
    queued: number;
    dropped: number;
    processed: number;
  }>("/scanner/status");
  const q = useQuery<{
    settings: Config;
    hasAPIKey: boolean;
    hasHeaders: boolean;
  }>("/settings");
  const [value, setValue] = useState<Config | null>(null);
  const [domains, setDomains] = useState("");
  const [suffixes, setSuffixes] = useState("");
  const [keywords, setKeywords] = useState("");
  const [dirty, setDirty] = useState(false);
  const [clearKey, setClearKey] = useState(false);
  const [clearHeaders, setClearHeaders] = useState(false);
  useUnsavedChanges(dirty);
  const [key, setKey] = useState("");
  const [headers, setHeaders] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  useEffect(() => {
    if (q.data && !dirty) {
      setValue(q.data.settings);
      setDomains((q.data.settings.domains || []).join("\n"));
      setSuffixes((q.data.settings.skipSuffixes || []).join("\n"));
      setKeywords((q.data.settings.denyKeywords || []).join("\n"));
    }
  }, [q.data, dirty]);
  const update = (v: Partial<Config>) => {
    setDirty(true);
    setValue((prev) => (prev ? { ...prev, ...v } : prev));
  };
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">
            {scanner ? "CAPTURE & SCOPE" : "CONFIGURATION"}
          </span>
          <h1>{scanner ? "扫描控制" : "系统设置"}</h1>
          <p>
            {scanner
              ? "限定目标范围，管理被动流量检测。"
              : "配置模型与过滤规则，敏感字段仅在本地保存。"}
          </p>
        </div>
        <div className="toolbar config-reload">
          <button
            disabled={busy}
            onClick={() => {
              if (dirty && !confirm("重新加载会放弃未保存修改，确定继续？"))
                return;
              setDirty(false);
              setValue(null);
              setKey("");
              setHeaders("");
              setClearKey(false);
              setClearHeaders(false);
              setError("");
              q.reload();
            }}
          >
            重新加载配置
          </button>
        </div>
      </div>
      <QueryState {...q} retry={q.reload} />
      {value && (
        <form
          className="configuration-form"
          onChange={() => setDirty(true)}
          onSubmit={async (e) => {
            e.preventDefault();
            setBusy(true);
            setError("");
            setMessage("");
            try {
              const lines = (s: string) =>
                s
                  .split("\n")
                  .map((v) => v.trim())
                  .filter(Boolean);
              const payload: Record<string, unknown> = {
                ...value,
                domains: lines(domains),
                skipSuffixes: lines(suffixes),
                denyKeywords: lines(keywords),
              };
              if (key) payload.apiKey = key;
              if (headers) {
                const parsed = JSON.parse(headers);
                if (
                  !parsed ||
                  Array.isArray(parsed) ||
                  typeof parsed !== "object" ||
                  Object.values(parsed).some((v) => typeof v !== "string")
                )
                  throw new Error("请求头必须是名称到文本值的 JSON 对象");
                payload.headers = parsed;
              }
              if (clearKey) payload.apiKey = "";
              if (clearHeaders) payload.headers = {};
              await api("/settings", send("PUT", payload));
              setDirty(false);
              setClearKey(false);
              setClearHeaders(false);
              setKey("");
              setHeaders("");
              setMessage("设置已保存");
              q.reload();
            } catch (e) {
              setError((e as Error).message);
            } finally {
              setBusy(false);
            }
          }}
        >
          <div className="configuration-main">
            {scanner ? (
              <>
                <section className="config-card">
                  <div className="config-card-heading">
                    <span className="config-icon">◎</span>
                    <div>
                      <h2>采集运行状态</h2>
                      <p>控制被动检测，暂停后代理仍可转发流量。</p>
                    </div>
                    <span
                      className={
                        "badge " + (value.paused ? "" : "badge-resolved")
                      }
                    >
                      {value.paused ? "已暂停" : "启用中"}
                    </span>
                  </div>
                  <div className="runtime-metrics">
                    {[
                      ["本次已处理", runtime.data?.processed || 0],
                      ["等待处理", runtime.data?.queued || 0],
                      ["队列满丢弃", runtime.data?.dropped || 0],
                    ].map(([label, n]) => (
                      <div key={label}>
                        <span>{label}</span>
                        <b>{n}</b>
                      </div>
                    ))}
                  </div>
                  <div className="config-row">
                    <label className="switch-label">
                      <input
                        type="checkbox"
                        checked={!value.paused}
                        onChange={(e) => update({ paused: !e.target.checked })}
                      />
                      <span>启用扫描</span>
                    </label>
                    <button
                      type="button"
                      className="subtle-button"
                      onClick={runtime.reload}
                    >
                      刷新运行状态
                    </button>
                  </div>
                </section>
                <section className="config-card">
                  <div className="config-card-heading">
                    <span className="config-icon">⌖</span>
                    <div>
                      <h2>目标范围</h2>
                      <p>只有范围内的流量会进入检测队列。</p>
                    </div>
                    <span className="count-pill">
                      {domains.split("\n").filter((v) => v.trim()).length}{" "}
                      个目标
                    </span>
                  </div>
                  <label className="config-field">
                    目标域名（每行一个精确域名或 IP）
                    <textarea
                      className="code-input"
                      rows={4}
                      value={domains}
                      onChange={(e) => setDomains(e.target.value)}
                      placeholder={"api.example.com\n10.0.0.10"}
                    />
                    <small>输入主机名，不包含协议、路径或通配符。</small>
                  </label>
                </section>
                <section className="config-card">
                  <div className="config-card-heading">
                    <span className="config-icon">♙</span>
                    <div>
                      <h2>测试身份 · 账号 B</h2>
                      <p>重放请求时替换凭证，与账号 A 的响应进行对比。</p>
                    </div>
                    <span
                      className={
                        "badge " + (q.data?.hasHeaders ? "badge-resolved" : "")
                      }
                    >
                      {q.data?.hasHeaders ? "已配置" : "待配置"}
                    </span>
                  </div>
                  <label className="config-field">
                    账号 B 请求头（JSON）
                    <textarea
                      className="code-input"
                      rows={4}
                      spellCheck={false}
                      value={headers}
                      onChange={(e) => setHeaders(e.target.value)}
                      placeholder={
                        q.data?.hasHeaders
                          ? "已有凭证已隐藏；留空保留当前值"
                          : '{\n  "Authorization": "Bearer …"\n}'
                      }
                    />
                    <small>
                      凭证不会以明文回显。留空保留当前值，输入新内容则整体替换。
                    </small>
                  </label>
                  <details className="advanced-options">
                    <summary>凭证管理</summary>
                    <label className="checkbox">
                      <input
                        type="checkbox"
                        checked={clearHeaders}
                        onChange={(e) => setClearHeaders(e.target.checked)}
                      />
                      保存时清除已有账号 B 请求头
                    </label>
                  </details>
                </section>
              </>
            ) : (
              <>
                <section className="config-card">
                  <div className="config-card-heading">
                    <span className="config-icon">✧</span>
                    <div>
                      <h2>模型连接</h2>
                      <p>使用兼容 Chat Completions 的模型辅助分析。</p>
                    </div>
                    <span
                      className={
                        "badge " + (q.data?.hasAPIKey ? "badge-resolved" : "")
                      }
                    >
                      {q.data?.hasAPIKey ? "密钥已配置" : "密钥待配置"}
                    </span>
                  </div>
                  <label className="config-field">
                    模型端点
                    <input
                      type="url"
                      value={value.endpoint}
                      onChange={(e) => update({ endpoint: e.target.value })}
                      placeholder="https://provider.example/v1/chat/completions"
                    />
                    <small>填写完整接口地址，而非供应商首页。</small>
                  </label>
                  <div className="config-field-grid">
                    <label className="config-field">
                      模型名称
                      <input
                        value={value.model}
                        onChange={(e) => update({ model: e.target.value })}
                        placeholder="输入实际模型 ID"
                      />
                    </label>
                    <label className="config-field">
                      API Key
                      <input
                        type="password"
                        autoComplete="off"
                        value={key}
                        onChange={(e) => setKey(e.target.value)}
                        placeholder={
                          q.data?.hasAPIKey
                            ? "留空保留当前密钥"
                            : "输入供应商密钥"
                        }
                      />
                    </label>
                  </div>
                  <details className="advanced-options">
                    <summary>密钥管理</summary>
                    <label className="checkbox">
                      <input
                        type="checkbox"
                        checked={clearKey}
                        onChange={(e) => setClearKey(e.target.checked)}
                      />
                      保存时清除已有模型密钥
                    </label>
                    <small>清除后可继续采集证据，通过人工复核完成判断。</small>
                  </details>
                </section>
                <section className="config-card">
                  <div className="config-card-heading">
                    <span className="config-icon">⌕</span>
                    <div>
                      <h2>检测过滤规则</h2>
                      <p>减少无关流量，让检测聚焦业务接口。</p>
                    </div>
                  </div>
                  <div className="config-field-grid">
                    <label className="config-field">
                      过滤后缀（每行一个）
                      <textarea
                        className="code-input"
                        rows={6}
                        value={suffixes}
                        onChange={(e) => setSuffixes(e.target.value)}
                      />
                      <small>匹配 URL 路径后缀；匹配到的流量不检测。</small>
                    </label>
                    <label className="config-field">
                      权限拒绝关键词（每行一个）
                      <textarea
                        rows={6}
                        value={keywords}
                        onChange={(e) => setKeywords(e.target.value)}
                        placeholder={"权限不足\n无访问权限"}
                      />
                      <small>命中后保留证据供人工核实，不直接确认安全。</small>
                    </label>
                  </div>
                </section>
              </>
            )}
          </div>
          <div className="configuration-rail">
            <section className="help-card">
              <span className="eyebrow">
                {scanner ? "CONNECTION GUIDE" : "ANALYSIS GUIDE"}
              </span>
              <h3>{scanner ? "连接你的测试流量" : "让模型辅助判断"}</h3>
              {scanner ? (
                <>
                  <ol>
                    <li>
                      <b>配置范围与身份</b>
                      <p>限定目标，并提供账号 B 的测试凭证。</p>
                    </li>
                    <li>
                      <b>连接上游代理</b>
                      <p>在 Burp 或浏览器代理设置中填写下方地址。</p>
                      <code>127.0.0.1:9080</code>
                    </li>
                    <li>
                      <b>保存并启用</b>
                      <p>访问测试接口，在检测结果中复核证据。</p>
                    </li>
                  </ol>
                  <div className="help-note">
                    HTTPS 流量采集需要信任本机代理证书。
                  </div>
                </>
              ) : (
                <>
                  <p>
                    模型收到脱敏后的对比证据。模型结论作为线索，需结合业务归属进行人工复核。
                  </p>
                  <div className="help-note">
                    未配置模型时仍可采集和对比证据。
                  </div>
                </>
              )}
            </section>
            <section className="help-card neutral">
              <h3>配置保存规则</h3>
              <p>保存后生效；未保存的更改会在离开页面时提醒。</p>
              <div className="help-rule">
                <span>凭证回显</span>
                <b>已隐藏</b>
              </div>
              <div className="help-rule">
                <span>并发保护</span>
                <b>版本校验</b>
              </div>
              <small>配置版本 {value.revision}</small>
            </section>
          </div>
          <div className="configuration-feedback">
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
          </div>
          <div className="configuration-savebar">
            <div className="draft-status">
              <small>{dirty ? "有未保存的修改" : "所有修改已保存"}</small>
            </div>
            <button className="primary" disabled={busy || q.loading}>
              {busy ? "保存中…" : "保存设置"}
            </button>
          </div>
        </form>
      )}
    </>
  );
}
