import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiClientError, api, bootstrapToken } from "./client";

describe("API 客户端", () => {
  afterEach(() => {
    sessionStorage.clear();
    window.history.replaceState(null, "", "/");
    vi.unstubAllGlobals();
  });

  it("从地址片段读取令牌并立即移除", () => {
    window.history.replaceState(null, "", "/#token=abc123");

    bootstrapToken();

    expect(sessionStorage.getItem("classpilot-token")).toBe("abc123");
    expect(window.location.hash).toBe("");
  });

  it("把 API 错误映射为可展示对象", async () => {
    sessionStorage.setItem("classpilot-token", "secret");
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: { code: "STUDENT_NUMBER_EXISTS", message: "同一班级内学号已存在" }
          }),
          { status: 409, headers: { "Content-Type": "application/json" } }
        )
      )
    );

    await expect(api.get("/students")).rejects.toEqual(
      new ApiClientError("STUDENT_NUMBER_EXISTS", "同一班级内学号已存在", 409)
    );
  });
});
