// dock.test.ts — the action dock's request store (ADR 0111 Delivery
// PR 3, #1958). Requests are pushed and closed like modal layers; the
// dock draws the strongest by rank, newest first on a tie.

import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";

import {
  _resetForTests,
  activeDockRequest,
  currentDockRequest,
  DOCK_RANKS,
  dockRequests,
  pushDockRequest,
  SHEET_HAND_WIDTH,
  SHEET_MAX_WIDTH,
  sheetWidth,
  takesBar,
  type DockRank,
  type DockRequest,
} from "./dock";

beforeEach(_resetForTests);

const req = (rank: DockRank, label: string): DockRequest => ({ rank, label });
const labels = () => get(dockRequests).map((r) => r.label);

describe("the dock request store", () => {
  it("is empty until something asks", () => {
    expect(get(activeDockRequest)).toBeNull();
    expect(currentDockRequest()).toBeNull();
  });

  it("push opens a request and close removes it, once", () => {
    const h = pushDockRequest(req("blocks", "declare blockers"));
    expect(get(activeDockRequest)?.label).toBe("declare blockers");
    h.close();
    expect(get(activeDockRequest)).toBeNull();
    // A second close (Svelte can run a cleanup twice) is harmless and
    // does not touch a request opened since.
    const other = pushDockRequest(req("step", "declare attackers"));
    h.close();
    expect(labels()).toEqual(["declare attackers"]);
    other.close();
  });

  it("orders by rank, strongest first, whatever order they were pushed in", () => {
    pushDockRequest(req("step", "attack row"));
    pushDockRequest(req("gameOver", "game over"));
    pushDockRequest(req("choice", "a may trigger"));
    pushDockRequest(req("blocks", "declare blockers"));
    pushDockRequest(req("flow", "a selection"));
    expect(labels()).toEqual([
      "a may trigger",
      "a selection",
      "declare blockers",
      "game over",
      "attack row",
    ]);
    expect(DOCK_RANKS).toEqual(["choice", "flow", "blocks", "gameOver", "step"]);
    expect(get(activeDockRequest)?.label).toBe("a may trigger");
  });

  it("breaks a tie by the newest, like a stack, and falls back when it closes", () => {
    const first = pushDockRequest(req("flow", "first"));
    const second = pushDockRequest(req("flow", "second"));
    expect(get(activeDockRequest)?.label).toBe("second");
    second.close();
    expect(get(activeDockRequest)?.label).toBe("first");
    first.close();
  });

  it("update replaces the content and keeps the request's place", () => {
    const older = pushDockRequest(req("flow", "older"));
    pushDockRequest(req("flow", "newer"));
    // A live count changing on the older one must not lift it above
    // the newer one.
    older.update({ ...req("flow", "older"), question: "2 declared" });
    expect(labels()).toEqual(["newer", "older"]);
    expect(get(dockRequests)[1]?.question).toBe("2 declared");
    // An update after close does not resurrect it.
    older.close();
    older.update(req("flow", "older"));
    expect(labels()).toEqual(["newer"]);
  });

  it("lets only a step row leave next and Pass turn in the bar", () => {
    expect(takesBar(null)).toBe(false);
    expect(takesBar(req("step", "attack row"))).toBe(false);
    for (const r of ["choice", "flow", "blocks", "gameOver"] as const) {
      expect(takesBar(req(r, r))).toBe(true);
    }
  });
});

describe("sheetWidth", () => {
  const sheet = (width?: number): DockRequest => ({
    rank: "flow",
    label: "x",
    sheet: { title: "x", width },
  });
  it("caps an ordinary picker at 720px", () => {
    expect(sheetWidth(sheet(2000))).toBe(SHEET_MAX_WIDTH);
    expect(sheetWidth(sheet(400))).toBe(400);
    expect(sheetWidth(sheet())).toBe(560);
  });
  it("lets the opening hand ask for the wide sheet (#2200)", () => {
    expect(sheetWidth(sheet(SHEET_HAND_WIDTH))).toBe(SHEET_HAND_WIDTH);
  });
});
