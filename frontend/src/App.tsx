import PasswordDialog from "./components/PasswordDialog";
import Overview, { type ResultSelection } from "./pages/Overview";
import Findings from "./pages/Findings";
import Users from "./pages/Users";
import Audit from "./pages/Audit";
import Settings from "./pages/Settings";
import { useEffect, useState } from "react";
import { api, send, setCSRF, type Session } from "./api/client";
import Login from "./pages/Login";
import Layout, { type Page } from "./components/Layout";
import useWorkspaceNavigation, {
  canLeaveWorkspace,
} from "./hooks/useWorkspaceNavigation";
export default function App() {
  const [session, setSession] = useState<Session | null>(null);
  const [setup, setSetup] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const {
    page,
    navigate: navigatePage,
    reset,
  } = useWorkspaceNavigation(session?.user.role);
  const [selection, setSelection] = useState<ResultSelection>({});
  const navigate = (next: Page, filter: ResultSelection = {}) => {
    if (!navigatePage(next)) return;
    setSelection(filter);
  };
  const [passwordOpen, setPasswordOpen] = useState(false);
  const accept = (s: Session, restore = false) => {
    setCSRF(s.csrf);
    setSession(s);
    if (!restore) reset();
  };
  useEffect(() => {
    const expire = () => {
      setSession(null);
      setCSRF("");
    };
    window.addEventListener("session-expired", expire);
    (async () => {
      try {
        const v = await api<{ setupRequired: boolean }>("/auth/status");
        setSetup(v.setupRequired);
        if (!v.setupRequired) {
          try {
            accept(await api<Session>("/auth/me"), true);
          } catch {}
        }
      } catch (e) {
        setError((e as Error).message);
      } finally {
        setLoading(false);
      }
    })();
    return () => window.removeEventListener("session-expired", expire);
  }, []);
  if (loading) return <div className="loading">正在连接工作台…</div>;
  if (error)
    return (
      <div className="loading">
        <p role="alert">{error}</p>
        <button onClick={() => location.reload()}>重试</button>
      </div>
    );
  if (!session)
    return (
      <Login setup={setup} onSetup={() => setSetup(false)} onLogin={accept} />
    );
  return (
    <Layout
      user={session.user}
      page={page}
      setPage={navigate}
      password={() => setPasswordOpen(true)}
      logout={async () => {
        if (!canLeaveWorkspace()) return;
        try {
          await api("/auth/logout", send("POST", {}));
          setSession(null);
          setCSRF("");
        } catch (e) {
          setError((e as Error).message);
        }
      }}
    >
      <section className="content">
        {page === "overview" ? (
          <Overview navigate={navigate} />
        ) : page === "findings" ? (
          <Findings
            key={JSON.stringify(selection)}
            initial={selection}
            admin={session.user.role === "admin"}
          />
        ) : page === "users" ? (
          <Users currentUser={session.user} />
        ) : page === "audit" ? (
          <Audit />
        ) : (
          <Settings key={page} scanner={page === "scanner"} />
        )}
      </section>
      {passwordOpen && (
        <PasswordDialog
          close={() => setPasswordOpen(false)}
          changed={() => {
            setPasswordOpen(false);
            setSession(null);
            setCSRF("");
          }}
        />
      )}
    </Layout>
  );
}
