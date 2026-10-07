import { useCallback, useEffect, useRef, useState } from "react";
import { navigation, type Page } from "../components/Layout";

export function canLeaveWorkspace() {
  return window.dispatchEvent(
    new Event("workspace-before-navigate", { cancelable: true }),
  );
}

function readPage(): Page {
  const key = window.location.hash.replace(/^#\/?/, "");
  return navigation.find(([page]) => page === key)?.[0] ?? "overview";
}

function writePage(page: Page, replace = false) {
  const url = `${window.location.pathname}${window.location.search}#/${page}`;
  window.history[replace ? "replaceState" : "pushState"](null, "", url);
}

export default function useWorkspaceNavigation(role?: string) {
  const [page, setPage] = useState<Page>(readPage);
  const current = useRef(page);
  const allowed = useCallback(
    (next: Page): Page =>
      role === "viewer" && next !== "overview" && next !== "findings"
        ? "overview"
        : next,
    [role],
  );
  const move = useCallback((next: Page, replace = false) => {
    current.current = next;
    writePage(next, replace);
    setPage(next);
  }, []);
  const navigate = useCallback(
    (next: Page) => {
      if (next === current.current) return true;
      if (!canLeaveWorkspace()) return false;
      move(allowed(next));
      return true;
    },
    [allowed, move],
  );
  useEffect(() => {
    move(allowed(current.current), true);
    const onHistory = () => {
      const next = allowed(readPage());
      if (next !== current.current && !canLeaveWorkspace()) {
        writePage(current.current, true);
        return;
      }
      move(next, true);
    };
    window.addEventListener("hashchange", onHistory);
    return () => {
      window.removeEventListener("hashchange", onHistory);
    };
  }, [allowed, move]);
  useEffect(() => {
    document.title = `${navigation.find(([key]) => key === page)?.[1]} · PrivHunter`;
  }, [page]);
  return { page: allowed(page), navigate, reset: () => move("overview", true) };
}
