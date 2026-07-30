import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";

describe("App", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    sessionStorage.clear();
  });

  it("显示中文产品名称", () => {
    render(<App />);

    expect(
      screen.getByRole("heading", { name: "ClassPilot 班主任工作台" })
    ).toBeInTheDocument();
  });

  it("没有当前班级时引导创建班级", async () => {
    stubAPI({
      "GET /api/v1/classes/current": null
    });

    render(<App />);

    expect(await screen.findByText("创建第一个班级")).toBeInTheDocument();
    expect(screen.getByLabelText("班级名称")).toBeInTheDocument();
  });

  it("显示当前班级和学生名单入口", async () => {
    stubAPI({
      "GET /api/v1/classes/current": {
        id: "class-1",
        name: "七年级一班",
        stage: "初中",
        grade: "七年级",
        schoolYear: "2026-2027",
        status: "current",
        notes: "",
        createdAt: "2026-07-31T00:00:00Z",
        updatedAt: "2026-07-31T00:00:00Z"
      },
      "GET /api/v1/students?classId=class-1&page=1&pageSize=50": {
        items: [],
        total: 0,
        page: 1,
        pageSize: 50
      }
    });

    render(<App />);

    expect(await screen.findByRole("heading", { name: "七年级一班" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "学生名单" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "字段设置" })).toBeInTheDocument();
  });
});

function stubAPI(responses: Record<string, unknown>) {
  sessionStorage.setItem("classpilot-token", "secret");
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const path = typeof input === "string" ? input : input.toString();
      const key = `${init?.method ?? "GET"} ${path}`;
      if (!(key in responses)) {
        return new Response(
          JSON.stringify({ error: { code: "TEST_ROUTE_MISSING", message: key } }),
          { status: 500, headers: { "Content-Type": "application/json" } }
        );
      }
      return new Response(JSON.stringify({ data: responses[key] }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    })
  );
}
