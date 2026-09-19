import { describe, expect, it } from "vitest";
import type { PlayerView } from "./protocol";
import {
  UNDO_UNLIMITED,
  canManageTable,
  formatUndoCount,
  hasUndoBudget,
  isUnlimitedUndo,
} from "./tableSettings";

// ADR 0075 §2.2/§2.3. Two properties, and the first one is the bug
// this module exists to prevent: under an unlimited budget the server
// reports -1 remaining on every seat, so every naive `remaining <= 0`
// check greys the Undo button on the table with the most undos.

function seat(over: Partial<PlayerView> = {}): PlayerView {
  return { id: "p1", name: "Alice", ...over } as PlayerView;
}

describe("isUnlimitedUndo", () => {
  it("is true for -1 and below, false for a real budget", () => {
    expect(isUnlimitedUndo(UNDO_UNLIMITED)).toBe(true);
    expect(isUnlimitedUndo(-2)).toBe(true);
    expect(isUnlimitedUndo(0)).toBe(false);
    expect(isUnlimitedUndo(3)).toBe(false);
  });

  it("is false for a missing value — absent is not unlimited", () => {
    expect(isUnlimitedUndo(undefined)).toBe(false);
    expect(isUnlimitedUndo(null)).toBe(false);
  });
});

describe("formatUndoCount", () => {
  it("renders an unlimited budget as ∞ and anything else as its number", () => {
    expect(formatUndoCount(UNDO_UNLIMITED)).toBe("∞");
    expect(formatUndoCount(0)).toBe("0");
    expect(formatUndoCount(3)).toBe("3");
    expect(formatUndoCount(undefined)).toBe("0");
  });
});

describe("hasUndoBudget", () => {
  it("lets an unlimited table undo although it reports -1 remaining", () => {
    expect(hasUndoBudget(seat({ undos_remaining: UNDO_UNLIMITED }), false)).toBe(true);
  });

  it("refuses an exhausted seat and allows one with budget", () => {
    expect(hasUndoBudget(seat({ undos_remaining: 0 }), false)).toBe(false);
    expect(hasUndoBudget(seat({ undos_remaining: 1 }), false)).toBe(true);
  });

  it("lets the admin through regardless — they bypass the budget", () => {
    expect(hasUndoBudget(seat({ undos_remaining: 0 }), true)).toBe(true);
    expect(hasUndoBudget(null, true)).toBe(true);
  });

  it("refuses a viewer with no seat", () => {
    expect(hasUndoBudget(null, false)).toBe(false);
  });
});

describe("canManageTable", () => {
  it("is the host or the admin, and nobody else", () => {
    expect(canManageTable("player", seat({ is_host: true }))).toBe(true);
    expect(canManageTable("admin", null)).toBe(true);
    expect(canManageTable("player", seat({ is_host: false }))).toBe(false);
    expect(canManageTable("player", seat())).toBe(false);
    expect(canManageTable("spectator", seat({ is_host: true }))).toBe(false);
    expect(canManageTable("identified", null)).toBe(false);
    expect(canManageTable(undefined, null)).toBe(false);
  });
});
