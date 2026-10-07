# PrivHunterAI

基于 Go / Gin、React / TypeScript 和 SQLite 的本地权限检测工作台。通过被动代理采集 HTTP(S) 流量，替换账号 B 凭证进行对比，结合兼容 Chat Completions 的模型分析与人工复核形成证据记录。

原项目：[Ed1s0nZ/PrivHunterAI](https://github.com/Ed1s0nZ/PrivHunterAI)。旧文档保留在 [docs/legacy-readme.md](docs/legacy-readme.md)，其中旧启动方式不再适用。

## 快速启动

依赖：Go 1.25 或以上、Node.js 22.12 或以上、C 编译器（SQLite 驱动使用 CGO）。

```sh
cd frontend
npm ci
npm run build
cd ..
go build -o bin/privhunter ./cmd/privhunter
./bin/privhunter
```

访问 `http://127.0.0.1:8222`，首次设置管理员；密码至少 12 字节。系统无公开注册，管理员在用户管理中添加只读成员或其他管理员。

在扫描控制中填写精确目标域名 / IP 和账号 B 请求头，再启用扫描。上游代理连接 `127.0.0.1:9080`。HTTPS 采集需要安装本机生成的代理 CA，默认位于 `~/.mitmproxy/mitmproxy-ca-cert.pem`。只在授权测试目标与测试账户上采集。

模型端点使用完整 Chat Completions URL。可通过 `PRIVHUNTER_API_KEY` 环境变量覆盖本地模型密钥，避免写入数据库。未配置模型时保留账号 A/B 证据供人工复核；HTTP 401/403 标记为拒绝，关键词命中与普通相似响应不会直接当作确定结论。模型看到的是脱敏证据，身份信息缺失可能导致 unknown。

## 功能

- 首次初始化、登录/退出、密码哈希、哈希会话、CSRF 与管理员/只读角色。
- SQLite 持久化用户、设置、结果、复核备注与操作审计。
- 概览、分页搜索、判定/复核筛选、账号 A/B 请求和响应对比。
- 人工确认、误报、解决状态，记录删除，CSV/JSON 筛选导出。
- 扫描暂停与目标范围、静态后缀过滤、凭证替换、5 分钟重复抑制。
- 64 条有界队列、请求超时、模型有限重试、1 MiB 响应及解压上限。
- 证据中的认证头、敏感 JSON 字段和查询参数脱敏；CSV 公式注入防护。

## 目录

```text
cmd/privhunter/       启动与服务生命周期
internal/
  auth/              密码、会话与用户权限约束
  config/            旧配置的显式本地迁移
  domain/            领域数据结构
  httpapi/           Gin 路由、中间件与接口处理
  scanner/           被动代理、队列、重放和模型适配
  security/          统一脱敏规则
  store/             SQLite schema 与数据访问
frontend/src/
  api/               API 契约与请求客户端
  components/        布局与共享展示组件
  hooks/             数据加载状态
  pages/             React 业务页面
  styles/            基础、组件、页面与响应式样式
docs/                重构契约与历史说明
```

## 运行配置与迁移

```sh
./bin/privhunter -listen 127.0.0.1:8222 -proxy 127.0.0.1:9080 -database data/privhunter.db -frontend frontend/dist
# 仅显式执行时覆盖本地扫描设置，保留暂停状态
./bin/privhunter -import-config ./config.json
```

默认仅监听本机。HTTPS 反向代理部署时设置 `-secure-cookie`，保持同源访问，禁止把 9080 代理直接暴露到公网。初始化应在本机完成后再开放远程访问。

数据目录、SQLite、旧配置、环境文件、日志和导出均被 Git 排除。SQLite 设置中保存的密钥是本地明文，依赖文件权限保护；不会从设置 API 返回。需要清除密钥时可通过 API 传入空字符串；空白 UI 输入表示保留当前值。数据库备份必须作为敏感文件保管。

## 开发与验证

```sh
# 在项目根目录验证后端
go test ./...
go test -race ./...
go vet ./...
# 启动后端
go run ./cmd/privhunter
# 另一终端，Vite 将 /api 转发到 8222
cd frontend
npm run dev
# 在 frontend 目录构建与浏览器验收（独立测试数据库）
npm run build
npm run test:e2e
```

工作台页面使用 `#/findings` 等地址，刷新保留当前页面，支持浏览器返回和前进。前端开发依赖和生产依赖可在 `frontend` 目录执行 `npm audit` 检查；后端在项目根目录执行 `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` 检查可触达漏洞。

前端通过 `npm run build` 执行严格 TypeScript 检查。服务未启用流量正文日志。敏感数据脱敏依赖字段名，业务特有格式应扩展 `internal/security` 并添加样例测试。
