import { defineConfig } from "@playwright/test";
export default defineConfig({
  testDir: "tests",
  workers: 1,
  use: { baseURL: "http://127.0.0.1:18222" },
  webServer: {
    command: `go run ../cmd/privhunter -listen 127.0.0.1:18222 -proxy 127.0.0.1:19080 -database ../data/e2e-${Date.now()}.db -frontend dist`,
    url: "http://127.0.0.1:18222/api/auth/status",
    reuseExistingServer: false,
    timeout: 120000,
  },
});
