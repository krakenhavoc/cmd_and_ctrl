import { describe, expect, it } from "vitest";
import type { GameMeta, SeatInfo } from "./api";
import {
  agoLabel,
  inviteDone,
  inviteHeading,
  inviteTables,
  openSeats,
  tableChoiceLabel,
  togetherLine,
} from "./playedWith";
import type { Session } from "./session";
import type { Tablemate } from "./tablemates";

const USER = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const DAY = 86_400_000;
const NOW = Date.parse("2026-11-05T12:00:00Z");

function seats(n: number): SeatInfo[] {
  return Array.from({ length: n }, (_, i) => ({
    player_id: `p${i}`,
    name: `P${i}`,
    seat: i,
  })) as SeatInfo[];
}

function table(partial: Partial<GameMeta> = {}, taken = 1): GameMeta {
  return {
    id: "g1",
    name: "FNM",
    created_at: "2026-11-05T00:00:00Z",
    players: seats(taken),
    state: "lobby",
    is_creator: true,
    ...partial,
  } as GameMeta;
}

function sess(partial: Partial<Session["principal"]> = {}, rest: Partial<Session> = {}): Session {
  const expires = new Date(NOW + DAY).toISOString();
  return {
    token: "t",
    expiresAt: expires,
    principal: {
      role: "identified",
      user_id: USER,
      name: "Alice",
      issued_at: new Date(NOW).toISOString(),
      expires_at: expires,
      ...partial,
    },
    ...rest,
  } as Session;
}

function mate(partial: Partial<Tablemate> = {}): Tablemate {
  return {
    user_id: "b",
    display_name: "Bob",
    last_played_at: NOW - 2 * DAY,
    games_together: 3,
    last_table: "FNM",
    ...partial,
  };
}

describe("openSeats", () => {
  it("is four minus the seats taken, bots included, and never negative", () => {
    expect(openSeats(table({}, 1))).toBe(3);
    expect(openSeats(table({}, 4))).toBe(0);
    expect(openSeats(table({}, 6))).toBe(0);
  });
});

describe("inviteTables", () => {
  it("offers your lobby tables with an empty seat", () => {
    expect(inviteTables([table()], sess())).toHaveLength(1);
  });

  it("drops a full table, one in play, one that ended, and a practice table", () => {
    const games = [
      table({ id: "full" }, 4),
      table({ id: "active", state: "active" }),
      table({ id: "ended", state: "ended" }),
      table({ id: "practice", practice: true }),
      table({ id: "ok" }),
    ];
    expect(inviteTables(games, sess()).map((g) => g.id)).toEqual(["ok"]);
  });

  it("drops a table that is neither yours nor one you are seated at", () => {
    const theirs = table({ id: "theirs", is_creator: false });
    expect(inviteTables([theirs], sess())).toEqual([]);
    expect(inviteTables([theirs], sess({}, { gameID: "theirs" }))).toHaveLength(1);
  });

  it("offers nothing to a guest, the admin token or no session", () => {
    expect(inviteTables([table()], sess({ role: "player", user_id: undefined }))).toEqual([]);
    expect(inviteTables([table()], sess({ role: "admin", user_id: undefined }))).toEqual([]);
    expect(inviteTables([table()], null)).toEqual([]);
  });
});

describe("the words", () => {
  it("names a table choice with its fill", () => {
    expect(tableChoiceLabel(table({ name: "Friday Pod" }, 2))).toBe("Friday Pod · 2/4");
  });

  it("says how many tables, the latest, and when", () => {
    expect(togetherLine(mate(), NOW)).toBe("3 tables together · last: FNM · 2 days ago");
    expect(togetherLine(mate({ games_together: 1 }), NOW)).toContain("1 table together");
  });

  it("copes with a server that sends neither new field", () => {
    const old = mate({ games_together: undefined, last_table: undefined });
    expect(togetherLine(old, NOW)).toBe("2 days ago");
    expect(togetherLine(mate({ ...old, last_played_at: 0 }), NOW)).toBe("played together");
  });

  it("keeps the ago label coarse", () => {
    expect(agoLabel(NOW - 1000, NOW)).toBe("today");
    expect(agoLabel(NOW - DAY, NOW)).toBe("yesterday");
    expect(agoLabel(NOW - 45 * DAY, NOW)).toBe("a month ago");
    expect(agoLabel(NOW - 400 * DAY, NOW)).toBe("a year ago");
    expect(agoLabel(undefined, NOW)).toBe("");
  });

  it("asks and confirms in the delivery the server has", () => {
    expect(inviteHeading("Bob", true)).toContain("Discord DM");
    expect(inviteHeading("Bob", false)).toContain("Copy an invite link");
    expect(inviteDone("Bob", "FNM", true)).toBe("Invite to FNM sent to Bob");
    expect(inviteDone("Bob", "FNM", false)).toBe("Invite link for FNM copied for Bob");
  });
});
