// lfgPost.test.ts — ADR 0161: when the #lfg box can be ticked, what the
// table settings say about the post, and the one route.

import { afterEach, describe, expect, it, vi } from "vitest";

import type { GameMeta, SeatInfo } from "./api";
import { LFG_GUEST_HINT, lfgBlocked, lfgStatusLine, setLFG } from "./lfgPost";

function seat(over: Partial<SeatInfo> = {}): SeatInfo {
  return { player_id: "p1", name: "Alice", seat: 0, deck_uploaded: false, ...over };
}

function table(over: Partial<GameMeta> = {}): GameMeta {
  return {
    id: "g1",
    name: "Friday pod",
    created_at: "2026-10-13T20:00:00Z",
    state: "lobby",
    players: [seat({ is_host: true, discord_id: "501" })],
    ...over,
  } as GameMeta;
}

describe("lfgBlocked", () => {
  it("lets a Discord host post a lobby table with an empty seat", () => {
    expect(lfgBlocked(table())).toBeNull();
  });

  it("tells a guest host to sign in with Discord", () => {
    expect(lfgBlocked(table({ players: [seat({ is_host: true })] }))).toBe(LFG_GUEST_HINT);
  });

  it("refuses a full table, a started one and an archived one", () => {
    const four = [0, 1, 2, 3].map((n) =>
      seat({ player_id: `p${n}`, seat: n, is_host: n === 0, discord_id: "501" }),
    );
    expect(lfgBlocked(table({ players: four }))).toMatch(/seat/);
    expect(lfgBlocked(table({ state: "active" }))).toMatch(/waiting to start/);
    expect(lfgBlocked(table({ archived_at: "2026-10-13T21:00:00Z" }))).toMatch(/waiting to start/);
  });

  it("never blocks withdrawing a post", () => {
    expect(lfgBlocked(table({ lfg: true, players: [seat({ is_host: true })] }))).toBeNull();
    expect(lfgBlocked(table({ lfg: true, state: "active" }))).toBeNull();
  });

  it("leaves a table with no host seated to the server", () => {
    expect(lfgBlocked(table({ players: [] }))).toBeNull();
  });
});

describe("lfgStatusLine", () => {
  it("says nothing for a table that is not posted", () => {
    expect(lfgStatusLine(table())).toBeNull();
  });

  it("words each status", () => {
    expect(lfgStatusLine(table({ lfg: true, lfg_status: "posted" }))).toMatch(/^Posted in #lfg/);
    expect(lfgStatusLine(table({ lfg: true, lfg_status: "pending" }))).toMatch(/^Posting/);
    expect(lfgStatusLine(table({ lfg: true, lfg_status: "unavailable" }))).toMatch(/unavailable/);
    expect(lfgStatusLine(table({ lfg: true }))).toMatch(/^Posting/);
  });

  it("says the post is down once the game starts", () => {
    expect(lfgStatusLine(table({ lfg: true, lfg_status: "pending", state: "active" }))).toMatch(
      /comes down once the game starts/,
    );
  });
});

describe("setLFG", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("PUTs the choice and answers the meta", async () => {
    const fetchMock = vi.fn(
      async () =>
        new Response(JSON.stringify(table({ lfg: true, lfg_status: "pending" })), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const meta = await setLFG("g 1", true);
    expect(meta.lfg_status).toBe("pending");
    const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe("/games/g%201/lfg");
    expect(init.method).toBe("PUT");
    expect(JSON.parse(init.body as string)).toEqual({ enabled: true });
  });
});
