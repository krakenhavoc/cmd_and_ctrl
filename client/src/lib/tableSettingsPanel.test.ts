// The derivations behind the table-settings panel (ADR 0075 §2.5).
//
// These are the decisions the CLIENT makes: which control is locked,
// what a patch is allowed to carry, and which of the two spawn gates
// a surface is answering. The server re-checks every one of them —
// nothing here is a rule — so what is worth pinning is the places a
// naive reading gets it backwards: that -1 is a setting and not an
// error, that "ended" locks starting life just as "active" does, and
// that the spawn entry needs BOTH gates while the badge needs
// neither.

import { describe, expect, it } from "vitest";

import type { GameView, PlayerView, TableSettingsView } from "./protocol";
import {
  DEFAULT_TABLE_SETTINGS,
  MAX_COMMANDER_DAMAGE,
  MAX_STARTING_LIFE,
  UNDO_UNLIMITED,
  canSpawn,
  describeUndoLimit,
  settingsDiff,
  settingsPatchError,
  spawnZonesFor,
  spawningVisible,
  startingLifeLocked,
  tableSettingsOf,
} from "./tableSettings";

function settings(over: Partial<TableSettingsView> = {}): TableSettingsView {
  return { ...DEFAULT_TABLE_SETTINGS, ...over };
}

const hostSeat = { id: "p1", name: "Alice", is_host: true } as PlayerView;
const plainSeat = { id: "p2", name: "Bob" } as PlayerView;

describe("tableSettingsOf", () => {
  it("falls back to the defaults when a frame carries no settings", () => {
    // An older server, or a replay frame captured before S35. The
    // panel then describes what that table is actually doing, which
    // is the default, rather than rendering blanks.
    expect(tableSettingsOf(null)).toEqual(DEFAULT_TABLE_SETTINGS);
    expect(tableSettingsOf({} as GameView)).toEqual(DEFAULT_TABLE_SETTINGS);
  });

  it("prefers the frame's own settings", () => {
    const s = settings({ allow_spawn: true });
    expect(tableSettingsOf({ settings: s } as GameView)).toBe(s);
  });
});

describe("describeUndoLimit", () => {
  it("reads -1 and 0 as words, because neither is a count", () => {
    expect(describeUndoLimit(UNDO_UNLIMITED)).toContain("Unlimited");
    expect(describeUndoLimit(0)).toContain("off");
  });

  it("agrees with itself about singular and plural", () => {
    expect(describeUndoLimit(1)).toContain("One take-back");
    expect(describeUndoLimit(3)).toContain("3 take-backs");
  });
});

describe("startingLifeLocked", () => {
  it("is open only in the lobby", () => {
    expect(startingLifeLocked("lobby")).toBe(false);
    expect(startingLifeLocked("active")).toBe(true);
    // A finished table has nothing left to deal, and the server's
    // rule is "not lobby" rather than "active".
    expect(startingLifeLocked("ended")).toBe(true);
    expect(startingLifeLocked(undefined)).toBe(true);
  });
});

describe("settingsPatchError", () => {
  it("accepts an unlimited undo budget — -1 is a value, not a mistake", () => {
    expect(settingsPatchError({ undo_limit: UNDO_UNLIMITED }, "active")).toBeNull();
    expect(settingsPatchError({ undo_limit: 0 }, "active")).toBeNull();
  });

  it("refuses a budget below unlimited", () => {
    expect(settingsPatchError({ undo_limit: -2 }, "active")).not.toBeNull();
  });

  it("holds the server's ranges for life and commander damage", () => {
    expect(settingsPatchError({ starting_life: 0 }, "lobby")).not.toBeNull();
    expect(settingsPatchError({ starting_life: MAX_STARTING_LIFE + 1 }, "lobby")).not.toBeNull();
    expect(settingsPatchError({ starting_life: 40 }, "lobby")).toBeNull();
    expect(settingsPatchError({ commander_damage: 0 }, "lobby")).not.toBeNull();
    expect(
      settingsPatchError({ commander_damage: MAX_COMMANDER_DAMAGE + 1 }, "lobby"),
    ).not.toBeNull();
    expect(settingsPatchError({ commander_damage: 21 }, "lobby")).toBeNull();
  });

  it("refuses a starting-life change once the game has started (the server's 422)", () => {
    expect(settingsPatchError({ starting_life: 30 }, "active")).toMatch(/already started/i);
    expect(settingsPatchError({ starting_life: 30 }, "lobby")).toBeNull();
  });

  it("says nothing about the fields that are always legal", () => {
    expect(
      settingsPatchError({ bot_pace: "slow", undo_scope: "host_any", allow_spawn: true }, "active"),
    ).toBeNull();
  });
});

describe("settingsDiff", () => {
  it("drops fields that already hold the proposed value", () => {
    const now = settings({ undo_limit: 3, allow_spawn: true });
    expect(settingsDiff(now, { undo_limit: 3, allow_spawn: true })).toEqual({});
  });

  it("keeps only what actually changed", () => {
    const now = settings({ undo_limit: 3 });
    expect(settingsDiff(now, { undo_limit: 3, bot_pace: "slow" })).toEqual({ bot_pace: "slow" });
  });

  it("keeps a change to false — absent and false are different patches", () => {
    const now = settings({ allow_spawn: true });
    expect(settingsDiff(now, { allow_spawn: false })).toEqual({ allow_spawn: false });
  });

  it("keeps a change to 0, which is a real undo setting", () => {
    const now = settings({ undo_limit: 1 });
    expect(settingsDiff(now, { undo_limit: 0 })).toEqual({ undo_limit: 0 });
  });
});

describe("canSpawn", () => {
  const on = settings({ allow_spawn: true });
  const off = settings({ allow_spawn: false });

  it("needs BOTH gates — manager and the table's switch", () => {
    expect(canSpawn("player", hostSeat, on)).toBe(true);
    expect(canSpawn("admin", null, on)).toBe(true);
    // Host, switch off.
    expect(canSpawn("player", hostSeat, off)).toBe(false);
    // Switch on, not the host.
    expect(canSpawn("player", plainSeat, on)).toBe(false);
    expect(canSpawn("spectator", null, on)).toBe(false);
  });
});

describe("spawningVisible", () => {
  it("is the switch and nothing else — the badge is for the opponents", () => {
    expect(spawningVisible(settings({ allow_spawn: true }))).toBe(true);
    expect(spawningVisible(settings({ allow_spawn: false }))).toBe(false);
  });
});

describe("spawnZonesFor", () => {
  it("offers a token the battlefield only", () => {
    // CR 704.5d: a token anywhere else ceases to exist at the next
    // state-based action check, so the zone would appear to work and
    // then silently undo itself.
    expect(spawnZonesFor("token")).toEqual(["battlefield"]);
  });

  it("offers a card every zone the spawn route takes", () => {
    expect(spawnZonesFor("card")).toContain("hand");
    expect(spawnZonesFor("card")).toContain("command");
    expect(spawnZonesFor("card")).not.toContain("stack");
  });
});
