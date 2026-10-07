import { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
type State<T> = {
  path: string | null;
  data: T | null;
  error: string;
  loading: boolean;
};
export function useQuery<T>(path: string | null) {
  const [state, setState] = useState<State<T>>({
    path: null,
    data: null,
    error: "",
    loading: false,
  });
  const [version, setVersion] = useState(0);
  useEffect(() => {
    if (path === null) {
      setState({ path: null, data: null, error: "", loading: false });
      return;
    }
    let active = true;
    const controller = new AbortController();
    setState((prev) => ({
      path,
      data: prev.path === path ? prev.data : null,
      error: "",
      loading: true,
    }));
    api<T>(path, { signal: controller.signal })
      .then((data) => {
        if (active) setState({ path, data, error: "", loading: false });
      })
      .catch((e) => {
        if (active && e.name !== "AbortError")
          setState((prev) => ({ ...prev, error: e.message, loading: false }));
      });
    return () => {
      active = false;
      controller.abort();
    };
  }, [path, version]);
  const reload = useCallback(() => setVersion((v) => v + 1), []);
  const current =
    state.path === path
      ? state
      : { path, data: null, error: "", loading: path !== null };
  return {
    data: current.data,
    error: current.error,
    loading: current.loading,
    reload,
  };
}
