import { describe, expect, it } from "vitest";

import type { GameMeta, SeatInfo } from "./api";
import type { MyGame } from "./myGames";
import {
  contextLine,
  filterHistory,
  historyRows,
  moreCount,
  nowTables,
  seatLine,
  statusChip,
} from "./tablesPage";

const HOUR = 3_600_000;
const NOW = Date.parse("2026-10-13T12:00:00Z");

function seat(n: number, extra: Partial<SeatInfo> = {}): SeatInfo {
  return { player_id: `p${n}`, name: `P${n}`, seat: n, deck_uploaded: true, ...extra };
}

function meta(extra: Partial<GameMeta> = {}): GameMeta {
  return {
    id: "g",
    name: "T",
    created_at: new Date(NOW - 2 * HOUR).toISOString(),
    players: [],
    state: "lobby",
    ...extra,
  };
}

function mine(id: string, extra: Partial<MyGame> = {}): MyGame {
  return {
    id,
    name: id,
    state: "ended",
    seat: 0,
    winner_seat: null,
    created_at: NOW - HOUR,
    started_at: null,
    ended_at: null,
    archived_at: null,
    others: [],
    ...extra,
  };
}

describe("the Now card's head (ADR 0158 §1 item 2)", () => {
  it("names the status, with the turn once one has begun", () => {
    expect(statusChip(meta())).toEqual({ label: "Lobby", live: false });
    expect(statusChip(meta({ state: "active", turn: 7 })).label).toBe("In play · turn 7");
    expect(statusChip(meta({ state: "active" })).label).toBe("In play");
  });

  it("says who hosts, your seat, and how old a lobby table is", () => {
    const players = [seat(0, { is_host: true, display_name: "Slysloth" }), seat(1)];
    expect(contextLine(meta({ players }), { seat: 0 }, NOW)).toBe("you host · created 2h ago");
    expect(contextLine(meta({ players, state: "active" }), { seat: 1 }, NOW)).toBe(
      "Slysloth hosts · your seat 2",
    );
    expect(contextLine(meta({ is_creator: true }), { seat: null }, NOW)).toBe(
      "you created it · created 2h ago",
    );
  });

  it("gives a seat its deck in the lobby and its deck and life in play", () => {
    const name = (id?: string) => (id === "raid" ? "Raid" : "");
    expect(seatLine(meta(), seat(0, { deck_name: "Atraxa" }), name).text).toBe("Atraxa");
    expect(seatLine(meta(), seat(0, { deck_name: "Atraxa", deck_power: 3.5 }), name)).toEqual({
      text: "Atraxa",
      pending: false,
      power: 3.5,
    });
    expect(
      seatLine(
        meta({ state: "active" }),
        seat(0, { deck_name: "Atraxa", deck_power: 4, life: 31 }),
        name,
      ),
    ).toEqual({ text: "Atraxa · 31 life", pending: false, power: 4 });
    expect(seatLine(meta(), seat(0, { deck_uploaded: false }), name)).toEqual({
      text: "deck pending",
      pending: true,
      power: 0,
    });
    expect(
      seatLine(meta({ state: "active" }), seat(0, { deck_name: "Atraxa", life: 31 }), name).text,
    ).toBe("Atraxa · 31 life");
    expect(
      seatLine(meta(), seat(1, { is_bot: true, bot_tier: "heuristic", bot_deck: "raid" }), name)
        .text,
    ).toBe("heuristic · Raid");
    expect(
      seatLine(
        meta(),
        seat(2, { is_bot: true, bot_tier: "assisted", bot_deck: "raid", bot_own_key: true }),
        name,
      ).text,
    ).toBe("assisted (own key) · Raid");
  });

  it("keeps ended and archived tables out of Now", () => {
    const list = [
      meta({ id: "a" }),
      meta({ id: "b", state: "active" }),
      meta({ id: "c", state: "ended" }),
      meta({ id: "d", archived_at: "2026-10-01T00:00:00Z" }),
    ];
    expect(nowTables(list).map((g) => g.id)).toEqual(["a", "b"]);
  });
});

describe("History (ADR 0158 §1 item 3)", () => {
  const rows = historyRows(
    [
      mine("won", { winner_seat: 0, ended_at: NOW - 3 * HOUR }),
      mine("lost", {
        winner_seat: 1,
        ended_at: NOW - 2 * HOUR,
        others: [{ seat: 1, name: "Bob" }],
      }),
      mine("shelved", { state: "active", archived_at: NOW - HOUR }),
      mine("now", { state: "active" }),
    ],
    [
      meta({
        id: "listed",
        state: "ended",
        created_at: new Date(NOW - 5 * HOUR).toISOString(),
      }),
      meta({ id: "won", state: "ended" }),
    ],
    new Set(["now"]),
  );

  it("is GET /me/games minus Now, plus the ended tables listed, most recent first", () => {
    expect(rows.map((r) => r.id)).toEqual(["shelved", "lost", "won", "listed"]);
    expect(rows.map((r) => r.result.label)).toEqual(["Archived", "Bob won", "You won", "Ended"]);
  });

  it("filters", () => {
    expect(filterHistory(rows, "won").map((r) => r.id)).toEqual(["won"]);
    expect(filterHistory(rows, "friends").map((r) => r.id)).toEqual(["lost"]);
    expect(filterHistory(rows, "archived").map((r) => r.id)).toEqual(["shelved"]);
    expect(filterHistory(rows, "all")).toHaveLength(4);
  });

  it("shows twenty more, or what is left", () => {
    expect(moreCount(27, 6)).toBe(20);
    expect(moreCount(27, 26)).toBe(1);
    expect(moreCount(6, 6)).toBe(0);
  });
});
