import { test, expect } from "@playwright/test";
test.describe.configure({ mode: "serial" });
test("initialize, navigate, configure, manage access, export, change password", async ({
  page,
}) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "创建管理员" })).toBeVisible();
  await page.getByLabel("用户名").fill("testadmin");
  await page.getByLabel("密码", { exact: true }).fill("fictional-password-123");
  await page.getByRole("button", { name: "初始化工作台" }).click();
  await expect(page.getByRole("heading", { name: "欢迎回来" })).toBeVisible();
  await page.getByLabel("密码", { exact: true }).fill("fictional-password-123");
  await page.getByRole("button", { name: "登录工作台" }).click();
  await expect(
    page.getByRole("heading", { name: "安全态势概览" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "检测结果", exact: false }).click();
  await expect(page.getByText("暂无记录")).toBeVisible();
  await page.getByRole("button", { name: "系统设置", exact: false }).click();
  await expect(page.getByLabel("模型端点")).toBeVisible();
  await expect(page).toHaveURL(/#\/settings$/);
  await page.reload();
  await expect(page.getByLabel("模型端点")).toBeVisible();
  await page.goBack();
  await expect(page.getByText("暂无记录")).toBeVisible();
  await page.goForward();
  await expect(page.getByLabel("模型端点")).toBeVisible();
  await page.screenshot({
    path: "../data/settings-redesign.png",
    fullPage: true,
  });
  await page.getByRole("button", { name: "扫描控制", exact: false }).click();
  await expect(page.getByLabel("目标域名")).toBeVisible();
  await page.screenshot({
    path: "../data/scanner-redesign.png",
    fullPage: true,
  });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({
    path: "../data/scanner-redesign-mobile.png",
    fullPage: true,
  });
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.getByLabel("目标域名").fill("example.test");
  await page.getByLabel("目标域名").press("End");
  await page.getByLabel("目标域名").press("Enter");
  await page.getByLabel("目标域名").pressSequentially("api.example.test");
  await expect(page.getByLabel("目标域名")).toHaveValue(
    "example.test\napi.example.test",
  );

  page.once("dialog", (dialog) => dialog.dismiss());
  await page.goBack();
  await expect(page).toHaveURL(/#\/scanner$/);
  await expect(page.getByLabel("目标域名")).toHaveValue(
    "example.test\napi.example.test",
  );
  page.once("dialog", (dialog) => dialog.dismiss());
  await page.getByRole("button", { name: "系统设置", exact: false }).click();
  await expect(page.getByRole("heading", { name: "扫描控制" })).toBeVisible();
  await expect(page.getByLabel("目标域名")).toHaveValue(
    "example.test\napi.example.test",
  );

  await page
    .getByRole("textbox", { name: "账号 B 请求头" })
    .fill('{"Authorization":"Bearer fictional-fixture"}');
  await page.getByRole("button", { name: "保存设置" }).click();
  await expect(page.getByRole("status")).toHaveText("设置已保存");
  await page.getByRole("button", { name: "用户管理", exact: false }).click();
  await page.screenshot({ path: "../data/users-redesign.png", fullPage: true });
  await page.getByRole("button", { name: "添加成员", exact: false }).click();
  await page.getByLabel("用户名").fill("viewer");
  await page.getByLabel("初始密码").fill("fictional-viewer-123");
  await page.getByRole("button", { name: "创建用户" }).click();
  await expect(
    page.getByRole("cell").filter({ hasText: "viewer" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "审计日志", exact: false }).click();
  await expect(
    page.getByRole("cell").filter({ hasText: "create_user" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "检测结果", exact: false }).click();
  const download = page.waitForEvent("download");
  await page.getByRole("link", { name: "导出 CSV" }).click();
  expect((await download).suggestedFilename()).toBe("privhunter-results.csv");
  await page.getByRole("button", { name: "testadmin", exact: true }).click();
  await expect(page.getByRole("dialog", { name: "修改密码" })).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "testadmin", exact: true }),
  ).toBeFocused();
  await page.getByRole("button", { name: "testadmin", exact: true }).click();
  await page.getByLabel("当前密码").fill("fictional-password-123");
  await page.getByLabel("新密码").fill("fictional-changed-123");
  await page.getByRole("button", { name: "保存新密码" }).click();
  await expect(page.getByRole("heading", { name: "欢迎回来" })).toBeVisible();
  await page.getByLabel("用户名").fill("viewer");
  await page.getByLabel("密码", { exact: true }).fill("fictional-viewer-123");
  await page.getByRole("button", { name: "登录工作台" }).click();
  await expect(
    page.getByRole("heading", { name: "安全态势概览" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "用户管理", exact: false }),
  ).toHaveCount(0);
  await page.goto("/#/users");
  await expect(
    page.getByRole("heading", { name: "安全态势概览" }),
  ).toBeVisible();
  await expect(page).toHaveURL(/#\/overview$/);
  await page.goto("/#/findings");
  await expect(page.getByText("暂无记录")).toBeVisible();
  await page.getByRole("button", { name: "概览", exact: false }).click();
  await page.screenshot({ path: "../data/e2e-desktop.png", fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({ path: "../data/e2e-mobile.png", fullPage: true });
  const denied = await page.request.get("/api/users");
  expect(denied.status()).toBe(403);
  await page.getByRole("button", { name: "退出登录" }).click();
  await expect(page.getByRole("heading", { name: "欢迎回来" })).toBeVisible();
});

test("proxy evidence appears in UI, review and reload persistence", async ({
  page,
}) => {
  const { createServer, request } = await import("node:http");
  const target = createServer((req, res) => {
    res.setHeader("Content-Type", "application/json");
    res.statusCode =
      req.headers.authorization === "Bearer synthetic-b" ? 403 : 200;
    res.end(JSON.stringify({ id: 1, token: "synthetic-hidden" }));
  });
  await new Promise<void>((resolve) => target.listen(0, "127.0.0.1", resolve));
  try {
    await page.goto("/");
    await page.getByLabel("用户名").fill("testadmin");
    await page
      .getByLabel("密码", { exact: true })
      .fill("fictional-changed-123");
    await page.getByRole("button", { name: "登录工作台" }).click();
    await expect(
      page.getByRole("heading", { name: "安全态势概览" }),
    ).toBeVisible();
    await page.getByRole("button", { name: "扫描控制", exact: false }).click();
    await page.getByLabel("目标域名").fill("127.0.0.1");
    await page
      .getByRole("textbox", { name: "账号 B 请求头" })
      .fill('{"Authorization":"Bearer synthetic-b"}');
    await page.getByLabel("启用扫描").check();
    await page.getByRole("button", { name: "保存设置" }).click();
    await expect(page.getByRole("status")).toHaveText("设置已保存");
    const addr = target.address() as { port: number };
    await new Promise<void>((resolve, reject) => {
      const req = request(
        {
          hostname: "127.0.0.1",
          port: 19080,
          path: `http://127.0.0.1:${addr.port}/orders`,
          headers: { Authorization: "Bearer synthetic-a" },
        },
        (res) => {
          res.resume();
          res.on("end", resolve);
        },
      );
      req.on("error", reject);
      req.end();
    });
    await page.getByRole("button", { name: "检测结果", exact: false }).click();
    await expect
      .poll(async () => {
        await page.getByRole("button", { name: "刷新", exact: true }).click();
        return page.getByRole("button", { name: /GET.*orders/ }).count();
      })
      .toBe(1);
    await page.getByRole("button", { name: /GET.*orders/ }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
    await expect(
      page.getByText("[REDACTED]", { exact: false }).first(),
    ).toBeVisible();
    await expect(
      page.getByText("synthetic-hidden", { exact: false }),
    ).toHaveCount(0);
    await page
      .getByRole("dialog")
      .getByRole("combobox")
      .selectOption("confirmed");
    await page.getByLabel("备注").fill("已复核测试证据");
    await page.getByRole("button", { name: "保存复核" }).click();
    await expect(page.getByRole("dialog")).toHaveCount(0);
    await expect(
      page.getByRole("cell").filter({ hasText: "已确认" }),
    ).toBeVisible();
    await page.reload();
    await page.getByRole("button", { name: "检测结果", exact: false }).click();
    await expect(
      page.getByRole("cell").filter({ hasText: "已确认" }),
    ).toBeVisible();
    await page.screenshot({ path: "../data/e2e-findings.png", fullPage: true });
    await page
      .getByRole("button", { name: "概览", exact: false })
      .first()
      .click();
    await expect(
      page.getByRole("button", { name: /已确认风险/ }),
    ).toContainText("1");
    await page.getByRole("button", { name: /已确认风险/ }).click();
    await expect(
      page.getByRole("combobox", { name: "复核状态", exact: true }),
    ).toHaveValue("confirmed");
    await page.getByRole("checkbox", { name: "选择当前页全部记录" }).check();
    await page
      .getByRole("combobox", { name: "批量复核状态" })
      .selectOption("false_positive");
    await page.getByRole("button", { name: "应用复核状态" }).click();
    await expect(page.getByRole("status")).toHaveText("已更新 1 条记录");
    await page.getByRole("button", { name: "清除筛选条件" }).click();
    await expect(
      page.getByRole("cell").filter({ hasText: "误报" }),
    ).toBeVisible();
  } finally {
    await new Promise<void>((resolve) => target.close(() => resolve()));
  }
});
