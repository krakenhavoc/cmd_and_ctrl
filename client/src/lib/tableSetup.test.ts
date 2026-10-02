// ADR 0110 Delivery PR 7 on the client: who is offered the create
// form, what the last setup says, the order tablemates are offered in,
// and the wire calls behind them.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { applyLastSetup, createGame, fetchLastDeck, fetchMySetup } from "./api";
import { setSession, type Session } from "./session";
import {
  canCreateTables,
  orderTablemates,
  setupResultMessage,
  setupSummary,
  type TableSetup,
} from "./tableSetup";
import { canInviteTablemates, type Tablemate } from "./tablemates";

const USER = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const NIL = "00000000-0000-0000-0000-000000000000";
const GAME = "11111111-2222-3333-4444-555555555555";

function sess(
  role: Session["principal"]["role"],
  userID: string | undefined,
  extra: Partial<Session> = {},
): Session {
  return {
    token: `${role}-tok`,
    expiresAt: "2026-12-01T00:00:00Z",
    principal: {
      role,
      user_id: userID,
      issued_at: "2026-10-01T00:00:00Z",
      expires_at: "2026-12-01T00:00:00Z",
    },
    ...extra,
  };
}

describe("canCreateTables", () => {
  it("offers the create form to any signed-in person, seated, watching or not", () => {
    expect(canCreateTables(sess("identified", USER))).toBe(true);
    expect(canCreateTables(sess("player", USER, { gameID: GAME, playerID: "p" }))).toBe(true);
    expect(canCreateTables(sess("spectator", USER))).toBe(true);
  });

  it("offers it to the admin token", () => {
    expect(canCreateTables(sess("admin", NIL))).toBe(true);
  });

  it("does not offer it to a guest", () => {
    expect(canCreateTables(sess("player", NIL, { gameID: GAME, playerID: "p" }))).toBe(false);
    expect(canCreateTables(sess("player", undefined))).toBe(false);
    expect(canCreateTables(sess("spectator", NIL))).toBe(false);
    expect(canCreateTables(null)).toBe(false);
  });
});

describe("canInviteTablemates for the creator", () => {
  it("offers the picker to the creator before they sit down", () => {
    expect(canInviteTablemates(sess("identified", USER), undefined, GAME, true)).toBe(true);
    expect(canInviteTablemates(sess("identified", USER), undefined, GAME, false)).toBe(false);
  });

  it("still refuses the admin token and a guest, creator bit or not", () => {
    expect(canInviteTablemates(sess("admin", NIL), undefined, GAME, true)).toBe(false);
    expect(canInviteTablemates(sess("player", NIL), GAME, GAME, true)).toBe(false);
  });
});

const SETUP: TableSetup = {
  settings: {
    undo_limit: 3,
    undo_scope: "own",
    starting_life: 30,
    commander_damage: 21,
    bot_pace: "fast",
    allow_spawn: true,
  },
  bots: [
    { tier: "random", deck_id: "raid", name: "Robo" },
    { tier: "heuristic", deck_id: "", name: "Pasted" },
  ],
  tablemates: [USER],
};

describe("setupSummary", () => {
  it("names the bots, their decks, and the settings that differ from a new table", () => {
    const names: Record<string, string> = { raid: "Raid and Ransack" };
    expect(setupSummary(SETUP, (id) => names[id] ?? id)).toBe(
      "2 bots: Robo (random, Raid and Ransack), Pasted (heuristic) · starting life 30 · fast bots · spawning on",
    );
  });

  it("says when there are no bots, and stays quiet about default settings", () => {
    expect(
      setupSummary({
        settings: {
          starting_life: 40,
          commander_damage: 21,
          bot_pace: "normal",
          allow_spawn: false,
        },
        bots: [],
        tablemates: [],
      }),
    ).toBe("no bots");
  });
});

describe("setupResultMessage", () => {
  it("says what was applied and names every skipped part with its reason", () => {
    expect(
      setupResultMessage({
        settings: true,
        bots_added: 1,
        skipped: [{ name: "Smart", reason: 'the "strong" tier is not available on this server' }],
      }),
    ).toBe(
      'Last setup: table settings applied, 1 bot added. Skipped Smart: the "strong" tier is not available on this server.',
    );
  });

  it("is empty when there is nothing to say", () => {
    expect(setupResultMessage(undefined)).toBe("");
    expect(setupResultMessage({ settings: false, bots_added: 0, skipped: [] })).toBe("");
  });
});

describe("orderTablemates", () => {
  const mate = (id: string): Tablemate => ({ user_id: id, display_name: id, last_played_at: 1 });

  it("puts the people from the last table first, marked, and keeps each group's order", () => {
    const out = orderTablemates([mate("a"), mate("b"), mate("c"), mate("d")], ["c", "a"]);
    expect(out.map((o) => [o.mate.user_id, o.atLastTable])).toEqual([
      ["a", true],
      ["c", true],
      ["b", false],
      ["d", false],
    ]);
  });

  it("is the server's order when there is no last table", () => {
    expect(orderTablemates([mate("a"), mate("b")], null).map((o) => o.atLastTable)).toEqual([
      false,
      false,
    ]);
  });
});

// --- the wire ---------------------------------------------------------

let calls: Array<{ url: string; method: string; body: unknown }> = [];
let answer: unknown = {};

beforeEach(() => {
  calls = [];
  answer = {};
  setSession(sess("identified", USER));
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      calls.push({
        url: input,
        method: init.method ?? "GET",
        body: typeof init.body === "string" ? JSON.parse(init.body) : undefined,
      });
      return {
        ok: true,
        status: 200,
        statusText: "OK",
        json: async () => answer,
        clone() {
          return this;
        },
      } as unknown as Response;
    }),
  );
});
afterEach(() => {
  vi.unstubAllGlobals();
  setSession(null);
});

describe("the setup and last-deck calls", () => {
  it("creates a table plain, or with the last setup", async () => {
    answer = {
      id: GAME,
      name: "x",
      players: [],
      setup: { settings: true, bots_added: 0, skipped: [] },
    };
    await createGame("Friday");
    const res = await createGame("Friday", { setup: "last" });
    expect(calls.map((c) => [c.method, c.url, c.body])).toEqual([
      ["POST", "/games", { name: "Friday" }],
      ["POST", "/games", { name: "Friday", setup: "last" }],
    ]);
    expect(res.setup?.settings).toBe(true);
  });

  it("reads the last setup and applies it to a table", async () => {
    answer = { setup: SETUP, game_id: GAME, updated_at: 5 };
    expect((await fetchMySetup()).setup?.bots).toHaveLength(2);
    answer = { game: { id: GAME }, settings: true, bots_added: 2, skipped: [] };
    await applyLastSetup(GAME);
    expect(calls.map((c) => [c.method, c.url, c.body])).toEqual([
      ["GET", "/me/setup", undefined],
      ["POST", `/games/${GAME}/setup`, { from: "last" }],
    ]);
  });

  it("reads the last deck, or null", async () => {
    answer = { last_deck: { kind: "library", id: USER } };
    expect(await fetchLastDeck()).toEqual({ kind: "library", id: USER });
    answer = { last_deck: null };
    expect(await fetchLastDeck()).toBeNull();
  });
});
