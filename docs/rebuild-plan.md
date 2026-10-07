# PrivHunterAI 重构交付契约

## Workflow Gate Report
- User request: 重构目录与前端，增加侧边栏、登录鉴权、SQLite，完善该场景功能，禁止上传敏感信息。
- Detected phase: P0/P1，现有实现无产品/API/安全设计文档。
- Task type: 已授权全栈重构和功能扩展。
- Required upstream artifacts: 功能范围、页面状态、数据模型、API、安全与迁移约定。
- Found artifacts: README、Go 代理与扫描实现、单页 HTML；截图仅展示旧仓库结构。
- Missing/weak artifacts: 自动化测试、持久化、鉴权、生命周期与模块边界。
- Implementation allowed now: yes，先形成本文最小契约，再按模块逐步实现。
- Prework required: 基线检查、敏感配置本地隔离。
- Execution scope: 下述功能全部纳入此次交付。
- Acceptance criteria: 构建、单元/接口/竞态测试通过；浏览器完成初始化、登录、导航、结果和设置流程；重启数据保留。
- Risks/assumptions: 默认本机部署；无公开注册；首次管理员初始化；AI 输出为辅助判断，保留人工复核。

## Maintainability Gate Report
- Requested change: 跨前端、传输、存储与扫描的重构。
- Files/modules inspected: main.go, index.go, scan.go, tools.go, mitmproxy.go, config, AIAPIS, index.html。
- Trigger: 多职责全局状态，HTML 814 行，跨层耦合。
- Current risk level: high。
- Responsibility count: HTTP、UI、配置、模型调用、扫描、导出、状态存储混合。
- Size/complexity signals: O(n*m) 相似度矩阵；无响应上限与请求超时。
- Coupling signals: 全局 Resp/logs；配置在 init 读取；通过包级变量通信。
- Tests covering the area: 无。
- Refactor required first: yes。
- Allowed change type: feature_after_refactor；用户明确授权整体重构与扩展。
- Proposed slice: 领域与存储、鉴权、HTTP/UI、扫描集成、端到端验证。
- Acceptance criteria: 不保留 AIAPIS 专用目录；无全局可变结果切片；服务与扫描共享明确存储接口。
- Validation commands: go test ./..., go test -race ./..., go vet ./..., go build ./cmd/privhunter, git diff --check。
- Risks and assumptions: 旧结果没有磁盘数据可迁移；旧配置只在本地迁移，不写入示例或文档。

## 功能与页面
1. 首次初始化管理员，登录、退出、修改密码，管理员/只读用户管理，禁用用户使会话失效。
2. SQLite 保存用户、哈希会话、结果、复核状态、运行设置与审计；事务迁移、外键、WAL、权限受限的本地数据目录。
3. 侧边栏：概览、检测结果、扫描控制、系统设置、用户管理、审计日志。
4. 概览统计与运行状态；结果分页、搜索、判定筛选、详情、请求/响应对比、人工复核、备注、删除、脱敏 CSV/JSON 导出。
5. 扫描暂停/恢复、目标域名范围、重复请求抑制、静态内容过滤、有界队列、超时、有限重试与错误记录；保留被动代理与凭证替换检测。
6. 通用兼容模型客户端，端点/模型可配置，密钥从环境或本地配置读取；不在 UI 返回明文密钥。
7. 设置保存与校验；空白、加载、失败、权限和窄屏状态完整；静态资源本地提供。
8. 默认脱敏凭证、敏感 JSON 字段与查询参数；日志不包含流量正文、密钥、Cookie 或模型完整错误响应。

## 结构与接口约定
- cmd/privhunter: 入口与信号退出。
- internal/config: 配置装载与校验。
- internal/domain: 公共领域类型。
- internal/store: SQLite、迁移与查询。
- internal/auth: 密码和会话。
- internal/httpapi: 路由、权限、请求校验、导出。
- internal/scanner: 被动代理、有界工作队列、比较与模型适配。
- internal/security: 脱敏。
- frontend/: React + TypeScript + Vite，按 pages/components/api/hooks 分层。
- frontend/dist/: Vite 构建产物，由 Gin 同源提供；启动参数允许单独替换前端构建目录。
- docs: 架构、运行和验证记录。

API 使用 /api 前缀，JSON 错误采用 {"error":"可读信息"}。会话 HttpOnly、SameSite=Strict，写入需要同源 CSRF token；非登录接口必须鉴权，配置/用户/扫描控制需要管理员。结果列表与导出共享筛选逻辑。浏览器通过文本节点显示流量，不插入未经转义的 HTML。

## 敏感数据约束
本地 config.json、.env、data/、SQLite 文件、证书、导出和运行日志不提交。已跟踪的旧 config.json 先本地隔离再从索引移除。示例仅包含占位符。此次不推送、不发布；提交前检查暂存区，扫描仅输出命中位置、不输出值。
