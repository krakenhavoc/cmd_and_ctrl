import { describe, expect, it } from "vitest";

import { BUG_DESC_MAX, BUG_TITLE_MAX, buildBugContext, validateBugReport } from "./bugReport";
import type { GameView } from "./protocol";

function viewWithTurn(turn: Partial<GameView["turn"]>): GameView {
  return {
    turn: {
      number: 0,
      active_seat: 0,
      priority_holder: 0,
      phase: "",
      step: "",
      ...turn,
    },
  } as GameView;
}

describe("buildBugContext", () => {
  it("degrades to id + connection when the view is null", () => {
    const ctx = buildBugContext("g-123", null, 0, "reconnecting");
    expect(ctx).toEqual({ game_id: "g-123", connection: "reconnecting" });
  });

  it("captures turn/phase/step/seq from a live view", () => {
    const view = viewWithTurn({ number: 7, phase: "combat", step: "declare_blockers" });
    const ctx = buildBugContext("g-1", view, 412, "connected");
    expect(ctx).toEqual({
      game_id: "g-1",
      connection: "connected",
      seq: 412,
      turn: 7,
      phase: "combat",
      step: "declare_blockers",
    });
  });

  it("omits seq when the watermark is still zero", () => {
    const ctx = buildBugContext("g-1", null, 0, "connected");
    expect(ctx.seq).toBeUndefined();
  });
});

describe("validateBugReport", () => {
  it("requires a non-blank title", () => {
    expect(validateBugReport("   ", "")).toMatch(/title is required/);
    expect(validateBugReport("it broke", "")).toBeNull();
  });

  it("accepts titles exactly at the cap and rejects one past it", () => {
    expect(validateBugReport("t".repeat(BUG_TITLE_MAX), "")).toBeNull();
    expect(validateBugReport("t".repeat(BUG_TITLE_MAX + 1), "")).toMatch(/too long/);
  });

  it("bounds the description", () => {
    expect(validateBugReport("ok", "d".repeat(BUG_DESC_MAX))).toBeNull();
    expect(validateBugReport("ok", "d".repeat(BUG_DESC_MAX + 1))).toMatch(/too long/);
  });
});
