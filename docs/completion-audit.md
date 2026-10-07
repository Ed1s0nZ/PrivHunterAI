# 最佳实践优化完成核对

2026-10-07。按用户的 React/Gin 重构、维护性、功能深化、管理页反馈与敏感信息本地保留要求，以及既有 `rebuild-plan.md` 的八项交付契约核对。未把「最佳实践」解释成所有未来功能或所有部署场景。

| 要求 | 当前状态证据 | 结论 |
| --- | --- | --- |
| 1. 初始化、登录、退出、改密、用户权限及会话失效 | auth/service.go、httpapi/server.go；认证测试及完整浏览器初始化/改密/只读流程；最后管理员测试 | 已完成 |
| 2. SQLite 持久化、会话哈希、迁移与受限权限 | schema.sql、migrations.go、store.go、auth/service.go；关闭重开测试、旧版本升级/新版本拒绝测试；本地目录 0700、库 0600 | 已完成 |
| 3. React 六页侧栏与 Gin 同源服务 | App/Layout、各 pages、package.json、cmd 入口和 Router；浏览器逐页导航、桌面和手机截图 | 已完成 |
| 4. 概览、结果搜索分页筛选、A/B 证据、复核备注、删除和导出 | dashboard/findings/detail/export/bulk API 与存储、Findings/EvidenceComparison；本地代理至 UI、复核刷新、指标跳转、批量复核、导出回归；摘要/证据/导出存储测试 | 已完成 |
| 5. 保留被动扫描，范围、暂停、去重、过滤、有界队列、超时、有限重试及失败记录 | scanner/engine.go、body.go、model.go；代理、凭证替换、头副本、去重、拒绝响应、模型 mock、解压上限测试 | 已完成 |
| 6. 通用模型适配、端点/名称配置、环境密钥、敏感字段不回显 | model.go、settings.go、Settings 页面；本地模型协议测试、API 密钥/请求头不返回测试 | 已完成 |
| 7. 设置校验及加载/空白/错误/权限/手机状态 | Settings/Status/useQuery、API 输入验证；多行输入、保存、拒绝离开草稿、页面隔离及权限浏览器流程；revision 并发冲突测试 | 已完成 |
| 8. 敏感凭证、JSON 和查询字段脱敏；正文与凭证不进日志 | security/redact.go、scanner 持久化与模型输入路径、通用错误文案；脱敏测试读取完整详情正文验证；本地配置凭证与可提交文件比对零命中 | 已完成 |
| 可维护目录、移除 AIAPIS、文档和运行入口 | cmd/internal 分职责、frontend api/components/hooks/pages/styles、Makefile/README；旧目录与 config 已从索引删除，本地 config 留存且忽略 | 已完成 |
| 用户反馈的登录割裂、管理页粗糙、页面简单 | 登录统一画布；配置分组/帮助侧栏/保存栏；成员摘要/弹窗；概览真实趋势与复核队列；桌面/手机实际截图和 E2E | 已完成 |
| 刷新/浏览器历史、详情性能、可维护样式 | 导航 hook 与权限回退；摘要和详情独立接口/类型；九个样式文件严格保留顺序，构建 CSS 拆分前后字节一致 | 已完成 |
| 不上传敏感信息 | config/env/data/db/cert/log/export 忽略，config 索引删除；没有执行 commit/push/发布，服务仅本地 | 已完成 |

## 本次最终检查

- 全仓库 `go test -race ./...`、`go vet ./...`、Go 二进制构建通过。
- React 严格 TypeScript/Vite 构建通过；样式拆分前后 CSS 哈希一致。
- 两项浏览器完整流程通过，覆盖配置与导航、用户/审计/导出/密码/权限，以及真实本地 HTTP 代理至证据复核。
- `git diff --check` 通过，运行页面 HTTP 200。
- `npm audit`（包含开发依赖）零漏洞。
- govulncheck 发现的三项可触达漏洞已通过 x/net v0.55.0、x/text v0.39.0 等升级消除；最低 Go 版本相应为 1.25，README 同步。复扫可触达漏洞为零。报告仍列出 1 项导入包、19 项模块级未触达公告，不能将结果表述为所有依赖零公告。

## 验证边界与后续建议

真实外部模型供应商没有调用；系统级代理 CA 安装和信任没有自动化验证。目前验证的是本地 HTTP 代理链路、模型兼容协议以及脱敏输入。字段脱敏规则不能推断业务自定义的任意敏感内容；相关说明见 verification.md。

experience-expansion.md 中明确列为后续建议的项目隔离、多身份、扫描任务历史、报告中心及模型用量统计，并非本轮确认的交付要求，仍未实现。原文的日期/方法过滤和连接测试也保持建议状态，不计作已交付功能。该审计不宣称未来功能全部完成。
