import { describe, expect, it } from "vitest";
import type { PlayerView } from "./protocol";
import {
  DEFAULT_TABLE_SETTINGS,
  UNDO_UNLIMITED,
  canManageTable,
  formatUndoCount,
  hasUndoBudget,
  isUnlimitedUndo,
  settingsSummary,
  undoBlockedReason,
} from "./tableSettings";

// #3499: the table card's rules line names only what differs.
describe("settingsSummary", () => {
  it("says Standard rules for the defaults", () => {
    expect(settingsSummary(DEFAULT_TABLE_SETTINGS)).toBe("Standard rules");
    expect(settingsSummary({ ...DEFAULT_TABLE_SETTINGS, free_mulligans: undefined })).toBe(
      "Standard rules",
    );
  });

  it("lists each change in a fixed order", () => {
    expect(
      settingsSummary({
        ...DEFAULT_TABLE_SETTINGS,
        allow_spawn: true,
        undo_limit: 0,
        starting_life: 30,
        bot_pace: "slow",
      }),
    ).toBe("30 life · no undos · slow pace · spawning on");
  });

  it("words the undo budget and mulligans", () => {
    const s = DEFAULT_TABLE_SETTINGS;
    expect(settingsSummary({ ...s, undo_limit: UNDO_UNLIMITED })).toBe("unlimited undos");
    expect(settingsSummary({ ...s, undo_limit: 3 })).toBe("3 undos a turn");
    expect(settingsSummary({ ...s, undo_scope: "host_any" })).toBe("host can undo anyone");
    expect(settingsSummary({ ...s, free_mulligans: 1 })).toBe("1 free mulligan");
    expect(settingsSummary({ ...s, free_mulligans: -1 })).toBe("every mulligan free");
    expect(settingsSummary({ ...s, commander_damage: 15 })).toBe("15 commander damage");
  });
});

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

// #2963: the dock offered "Undo 1" while the server refused the undo
// because the most recent action was not the viewer's.
describe("undoBlockedReason", () => {
  it("is null when the stack allows it and budget is left", () => {
    expect(undoBlockedReason(seat({ undos_remaining: 1 }), false)).toBeNull();
    expect(undoBlockedReason(seat({ undos_remaining: UNDO_UNLIMITED }), false)).toBeNull();
  });

  it("names the stack's reason even with budget left", () => {
    const nothing = undoBlockedReason(seat({ undos_remaining: 1, undo_blocked: "nothing" }), false);
    expect(nothing).toContain("nothing to undo");
    const theirs = undoBlockedReason(
      seat({ undos_remaining: 1, undo_blocked: "not_yours" }),
      false,
    );
    expect(theirs).toContain("only undo your own most recent action");
    const over = undoBlockedReason(seat({ undos_remaining: 1, undo_blocked: "game_over" }), false);
    expect(over).toContain("game is over");
  });

  it("names the spent budget when the stack allows it", () => {
    expect(undoBlockedReason(seat({ undos_remaining: 0 }), false)).toContain("no undos remaining");
  });

  it("never blocks an unseated admin", () => {
    expect(undoBlockedReason(null, true)).toBeNull();
    expect(undoBlockedReason(seat({ undo_blocked: "not_yours" }), true)).toBeNull();
  });

  // ADR 0194: a bot's move, and anything older than the last eight
  // actions, is refused for everyone, the admin included.
  it("names a barrier for everyone, admin included", () => {
    for (const admin of [false, true]) {
      expect(
        undoBlockedReason(seat({ undos_remaining: 1, undo_blocked: "past_bot" }), admin),
      ).toContain("a bot has acted since");
      expect(
        undoBlockedReason(seat({ undos_remaining: 1, undo_blocked: "too_deep" }), admin),
      ).toContain("only goes back 8 actions");
    }
  });
});
