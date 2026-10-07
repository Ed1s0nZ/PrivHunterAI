export type User = {
  id: number;
  username: string;
  role: "admin" | "viewer";
  disabled: boolean;
};
export type Session = { user: User; csrf: string };
export type Finding = {
  id: number;
  method: string;
  url: string;
  requestA: string;
  requestB: string;
  responseA: string;
  responseB: string;
  result: string;
  reason: string;
  confidence: string;
  review: string;
  note: string;
  createdAt: string;
};
export type FindingSummary = Omit<
  Finding,
  "requestA" | "requestB" | "responseA" | "responseB"
>;
let csrf = "";
export function setCSRF(value: string) {
  csrf = value;
}
export async function api<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const response = await fetch("/api" + path, {
    ...options,
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrf,
      ...options.headers,
    },
  });
  const body = await response.json();
  if (!response.ok) {
    if (response.status === 401)
      window.dispatchEvent(new Event("session-expired"));
    throw new Error(body.error || "请求失败");
  }
  return body as T;
}
export const send = (method: string, value: unknown): RequestInit => ({
  method,
  body: JSON.stringify(value),
});
