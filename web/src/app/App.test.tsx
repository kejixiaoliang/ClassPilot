import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { App } from "./App";

describe("App", () => {
  it("显示中文产品名称", () => {
    render(<App />);

    expect(
      screen.getByRole("heading", { name: "ClassPilot 班主任工作台" })
    ).toBeInTheDocument();
  });
});
