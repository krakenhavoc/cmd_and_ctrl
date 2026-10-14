// resultsPost.test.ts — ADR 0174: when the results box can be ticked,
// what the table settings say about it, and the one route.

import { afterEach, describe, expect, it, vi } from "vitest";

import type { GameMeta } from "./api";
import { resultsBlocked, resultsStatusLine, setResults } from "./resultsPost";

function table(over: Partial<GameMeta> = {}): GameMeta {
  return {
    id: "g1",
    name: "Friday pod",
    created_at: "2026-10-14T20:00:00Z",
    state: "lobby",
    players: [],
    ...over,
  } as GameMeta;
}

describe("resultsBlocked", () => {
  it("lets a host opt in a table that has not ended, guest host or not", () => {
    expect(resultsBlocked(table())).toBeNull();
    expect(resultsBlocked(table({ state: "active" }))).toBeNull();
  });

  it("refuses a practice table, an ended game and an archived one", () => {
    expect(resultsBlocked(table({ practice: true }))).toMatch(/practice/);
    expect(resultsBlocked(table({ state: "ended" }))).toMatch(/ended/);
    expect(resultsBlocked(table({ archived_at: "2026-10-14T21:00:00Z" }))).toMatch(/ended/);
  });

  it("never blocks opting out", () => {
    expect(resultsBlocked(table({ results: true, state: "ended" }))).toBeNull();
    expect(
      resultsBlocked(table({ results: true, archived_at: "2026-10-14T21:00:00Z" })),
    ).toBeNull();
  });
});

describe("resultsStatusLine", () => {
  it("says when the result will not be posted", () => {
    expect(resultsStatusLine(table())).toMatch(/will not be posted/);
  });

  it("says when it will, and the early-concession exception", () => {
    const line = resultsStatusLine(table({ results: true }));
    expect(line).toMatch(/posted to Discord when the game ends/);
    expect(line).toMatch(/conceded in its first turns/);
  });

  it("does not claim an ended game was posted", () => {
    expect(resultsStatusLine(table({ results: true, state: "ended" }))).toMatch(
      /posted if it was played out/,
    );
  });
});

describe("setResults", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("PUTs the choice and answers the meta", async () => {
    const fetchMock = vi.fn(
      async () =>
        new Response(JSON.stringify(table({ results: true })), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const meta = await setResults("g 1", true);
    expect(meta.results).toBe(true);
    const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe("/games/g%201/results");
    expect(init.method).toBe("PUT");
    expect(JSON.parse(init.body as string)).toEqual({ enabled: true });
  });
});
