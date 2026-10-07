import { useEffect } from "react";
export function useUnsavedChanges(dirty: boolean) {
  useEffect(() => {
    if (!dirty) return;
    const unload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
    };
    const navigate = (e: Event) => {
      if (!window.confirm("有未保存的修改，确定放弃并离开？"))
        e.preventDefault();
    };
    window.addEventListener("beforeunload", unload);
    window.addEventListener("workspace-before-navigate", navigate);
    return () => {
      window.removeEventListener("beforeunload", unload);
      window.removeEventListener("workspace-before-navigate", navigate);
    };
  }, [dirty]);
}
