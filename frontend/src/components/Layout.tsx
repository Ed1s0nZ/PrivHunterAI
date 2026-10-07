import type { ReactNode } from "react";
import type { User } from "../api/client";
export const navigation = [
  ["overview", "概览", "◈"],
  ["findings", "检测结果", "⌕"],
  ["scanner", "扫描控制", "◎"],
  ["settings", "系统设置", "⚙"],
  ["users", "用户管理", "♙"],
  ["audit", "审计日志", "≡"],
] as const;
export type Page = (typeof navigation)[number][0];
export default function Layout({
  user,
  page,
  setPage,
  logout,
  password,
  children,
}: {
  user: User;
  page: Page;
  setPage: (p: Page) => void;
  logout: () => void;
  password: () => void;
  children: ReactNode;
}) {
  return (
    <div className="shell">
      <aside>
        <div className="brand">
          <b>PH</b>
          <div>
            PrivHunter<span>安全分析工作台</span>
          </div>
        </div>
        <span className="nav-label">WORKSPACE</span>
        <nav>
          {navigation
            .filter(
              ([key]) =>
                user.role === "admin" ||
                key === "overview" ||
                key === "findings",
            )
            .map(([key, label, icon]) => (
              <button
                key={key}
                className={page === key ? "active" : ""}
                aria-current={page === key ? "page" : undefined}
                onClick={() => setPage(key)}
              >
                <span>{icon}</span>
                {label}
              </button>
            ))}
        </nav>
        <div className="sidebar-footer">
          <div className="avatar">
            {user.username.slice(0, 1).toUpperCase()}
          </div>
          <div>
            <button
              className="account-button"
              onClick={password}
              title="修改密码"
            >
              {user.username}
            </button>
            <small>{user.role === "admin" ? "管理员" : "只读成员"}</small>
          </div>
          <button onClick={logout} aria-label="退出登录">
            ↪
          </button>
        </div>
      </aside>
      <main>
        <header>
          <span>工作台 / {navigation.find(([key]) => key === page)?.[1]}</span>
          <span className="local-badge">● LOCAL WORKSPACE</span>
        </header>
        {children}
      </main>
    </div>
  );
}
