import { describe, expect, it } from "vitest";

import type { GameMeta } from "./api";
import { FORMATS, fitsFormat, formatOf, tableSeats } from "./formats";
import { lfgBlocked } from "./lfgPost";
import { openSeats, tableChoiceLabel } from "./playedWith";
import { DEFAULT_TABLE_SETTINGS, settingsSummary, settingsTopics } from "./tableSettings";
import { setupSummary } from "./tableSetup";
import { lobbySlots } from "./tablesPage";

// ADR 0222 §1: a format on the table, on the client. Each block also
// pins the Commander side: a table, deck or setup that names no format
// reads exactly as it did before formats.

function meta(over: Partial<GameMeta> = {}): GameMeta {
  return {
    id: "g1",
    name: "FNM",
    created_at: "2026-11-10T00:00:00Z",
    players: [],
    state: "lobby",
    ...over,
  } as GameMeta;
}

function seat(n: number): GameMeta["players"][number] {
  return { player_id: `p${n}`, name: `P${n}`, seat: n } as GameMeta["players"][number];
}

describe("formatOf", () => {
  it("reads nothing, an empty name and an unknown name as Commander", () => {
    for (const id of [undefined, null, "", "vintage"]) {
      expect(formatOf(id).id).toBe("commander");
    }
    expect(formatOf("standard").id).toBe("standard");
  });

  it("lists Commander first, with the server's numbers", () => {
    expect(FORMATS.map((f) => f.id)).toEqual(["commander", "standard"]);
    expect(formatOf("commander")).toMatchObject({ seats: 4, startingLife: 40, hasCommander: true });
    expect(formatOf("standard")).toMatchObject({ seats: 2, startingLife: 20, hasCommander: false });
    expect(formatOf("commander").name).toBe("Commander");
    expect(formatOf("standard").name).toBe("Standard");
  });

  it("keeps the Commander deck copy and gives Standard its own", () => {
    expect(formatOf(undefined).deckRules).toMatch(/Commander rules — 100-card singleton/);
    expect(formatOf(undefined).pastePlaceholder).toMatch(/Commander:\n1 Atraxa/);
    expect(formatOf("standard").deckRules).toMatch(
      /at least 60 cards.*at most 4.*sideboard of up to 15/,
    );
    expect(formatOf("standard").pastePlaceholder).not.toMatch(/Commander/);
  });
});

describe("fitsFormat", () => {
  it("offers a deck at tables of its own format only", () => {
    expect(fitsFormat(undefined, undefined)).toBe(true);
    expect(fitsFormat(undefined, "commander")).toBe(true);
    expect(fitsFormat("standard", "standard")).toBe(true);
    expect(fitsFormat(undefined, "standard")).toBe(false);
    expect(fitsFormat("standard", undefined)).toBe(false);
  });
});

describe("a table's seats", () => {
  it("is four at a table that names no format", () => {
    const g = meta({ players: [seat(0)] });
    expect(tableSeats(g)).toBe(4);
    expect(lobbySlots(g)).toHaveLength(4);
    expect(openSeats(g)).toBe(3);
    expect(tableChoiceLabel(g)).toBe("FNM · 1/4");
  });

  it("is two at a Standard table", () => {
    const g = meta({ format: "standard", max_seats: 2, players: [seat(0)] });
    expect(tableSeats(g)).toBe(2);
    expect(lobbySlots(g)).toHaveLength(2);
    expect(lobbySlots(g)[1]).toBeNull();
    expect(openSeats(g)).toBe(1);
    expect(tableChoiceLabel(g)).toBe("FNM · 1/2");
    // An older server that names the format but no cap: the format's.
    expect(tableSeats(meta({ format: "standard" }))).toBe(2);
  });

  it("is full for an #lfg post at two", () => {
    const full = meta({ format: "standard", max_seats: 2, players: [seat(0), seat(1)] });
    expect(lfgBlocked(full)).toBe("Every seat is taken.");
    const commander = meta({ players: [seat(0), seat(1)] });
    expect(lfgBlocked(commander)).not.toBe("Every seat is taken.");
  });
});

describe("the rules line", () => {
  it("is unchanged at a Commander table", () => {
    expect(settingsSummary(DEFAULT_TABLE_SETTINGS)).toBe("40 life · 1 undo · standard rules");
    expect(settingsSummary({ ...DEFAULT_TABLE_SETTINGS, commander_damage: 30 })).toBe(
      "40 life · 1 undo · 30 commander damage",
    );
  });

  it("leads with the format at a Standard table, and never says commander damage", () => {
    const s = { ...DEFAULT_TABLE_SETTINGS, format: "standard", starting_life: 20 };
    expect(settingsSummary(s)).toBe("Standard · 20 life · 1 undo");
    expect(settingsSummary({ ...s, commander_damage: 30, allow_spawn: true })).toBe(
      "Standard · 20 life · 1 undo · spawning on",
    );
  });

  it("names commander damage in the tooltip only where there are commanders", () => {
    expect(settingsTopics(true)).toBe(
      "undos, starting life, commander damage, table pace, spawning",
    );
    expect(settingsTopics(false)).toBe("undos, starting life, table pace, spawning");
    expect(settingsTopics(true, "life")).toBe(
      "undos, life, commander damage, table pace, spawning",
    );
  });
});

describe("a saved setup's line", () => {
  it("is unchanged for a setup that names no format", () => {
    expect(setupSummary({ settings: { starting_life: 40 }, bots: [], tablemates: [] })).toBe(
      "no bots",
    );
    expect(setupSummary({ settings: { starting_life: 30 }, bots: [], tablemates: [] })).toBe(
      "no bots · starting life 30",
    );
  });

  it("names a Standard setup and compares its life with Standard's", () => {
    expect(
      setupSummary({
        settings: { format: "standard", starting_life: 20 },
        bots: [],
        tablemates: [],
      }),
    ).toBe("Standard · no bots");
    expect(
      setupSummary({
        settings: { format: "standard", starting_life: 25, commander_damage: 30 },
        bots: [],
        tablemates: [],
      }),
    ).toBe("Standard · no bots · starting life 25");
  });
});
