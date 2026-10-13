// moderation.test.ts — the client half of #2984: who can be reported, the
// refusals the server answers with a stable code, the report and block
// requests it sends, and the Reports tab's hash.

import { afterEach, describe, expect, it, vi } from "vitest";

import { removeSeat, reportPlayer } from "./api";
import { parseHash } from "./router";
import { LobbyApiError } from "./session";
import { lobbyErrorMessage } from "./serverFull";
import { adminViewHash, REPORTS_HASH, tabOf } from "./adminViews";
import {
  BLOCK_HINT,
  BLOCKED_FROM_TABLE_CODE,
  canSendReport,
  isBlockedFromTable,
  isNameRefused,
  NAME_NOT_ALLOWED_CODE,
  reportErrorMessage,
  reportTargets,
} from "./moderation";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("reportTargets", () => {
  const seats = [
    { id: "me", name: "Me" },
    { id: "bob", name: "Bob" },
    { id: "bot", name: "Bot 1", is_bot: true },
    { id: "cat", name: "Cat" },
  ];

  it("lists every other human seat, and no bot", () => {
    expect(reportTargets(seats, "me")).toEqual([
      { id: "bob", name: "Bob" },
      { id: "cat", name: "Cat" },
    ]);
  });

  it("lists nobody for a viewer with no seat (a spectator, the unseated admin)", () => {
    expect(reportTargets(seats, null)).toEqual([]);
  });

  it("lists nobody when the only other seats are bots", () => {
    expect(
      reportTargets(
        seats.filter((s) => s.id === "me" || s.is_bot),
        "me",
      ),
    ).toEqual([]);
  });
});

describe("canSendReport", () => {
  it("needs a player and a reason that says something", () => {
    expect(canSendReport("bob", "slurs in chat")).toBe(true);
    expect(canSendReport("", "slurs in chat")).toBe(false);
    expect(canSendReport("bob", "   ")).toBe(false);
  });
});

describe("the server's two coded refusals", () => {
  const blocked = new LobbyApiError(403, "you were removed from this table", undefined, undefined, {
    error: "you were removed from this table",
    code: BLOCKED_FROM_TABLE_CODE,
  });
  const name = new LobbyApiError(
    400,
    "that name isn't allowed here; pick another",
    undefined,
    undefined,
    {
      code: NAME_NOT_ALLOWED_CODE,
    },
  );

  it("are told apart by code, not by wording", () => {
    expect(isBlockedFromTable(blocked)).toBe(true);
    expect(isBlockedFromTable(name)).toBe(false);
    expect(isNameRefused(name)).toBe(true);
    expect(isNameRefused(blocked)).toBe(false);
    expect(isBlockedFromTable(new Error("x"))).toBe(false);
    expect(isBlockedFromTable(new LobbyApiError(403, "forbidden"))).toBe(false);
  });

  it("reach the join form as the server's own sentence", () => {
    expect(lobbyErrorMessage(blocked, "join failed")).toBe("you were removed from this table");
    expect(lobbyErrorMessage(name, "join failed")).toBe(
      "that name isn't allowed here; pick another",
    );
  });
});

describe("reportErrorMessage", () => {
  it("gives the rate limit its own plain sentence", () => {
    expect(reportErrorMessage(new LobbyApiError(429, "too many requests"))).toMatch(
      /try again in a few minutes/i,
    );
  });
  it("shows the server's sentence for any other refusal", () => {
    expect(reportErrorMessage(new LobbyApiError(422, "you can't report your own seat"))).toBe(
      "you can't report your own seat",
    );
  });
  it("has a fallback for a network failure", () => {
    expect(reportErrorMessage(new TypeError("Failed to fetch"))).toMatch(/couldn't send/i);
  });
});

describe("the requests", () => {
  function stubFetch(): { calls: { url: string; init: RequestInit }[] } {
    const calls: { url: string; init: RequestInit }[] = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init: RequestInit) => {
        calls.push({ url, init });
        return new Response(JSON.stringify({ id: "r1", players: [] }), { status: 200 });
      }),
    );
    return { calls };
  }

  it("removeSeat sends no block by default and ?block=true on request", async () => {
    const { calls } = stubFetch();
    await removeSeat("g1", "p1");
    await removeSeat("g1", "p1", { block: false });
    await removeSeat("g1", "p1", { block: true });
    expect(calls.map((c) => c.url)).toEqual([
      "/games/g1/seats/p1",
      "/games/g1/seats/p1",
      "/games/g1/seats/p1?block=true",
    ]);
    expect(calls.every((c) => c.init.method === "DELETE")).toBe(true);
  });

  it("reportPlayer posts the seat and the reason to the table's reports route", async () => {
    const { calls } = stubFetch();
    const out = await reportPlayer("g1", "p2", "slurs in chat");
    expect(out.id).toBe("r1");
    expect(calls[0].url).toBe("/games/g1/reports");
    expect(calls[0].init.method).toBe("POST");
    expect(JSON.parse(calls[0].init.body as string)).toEqual({
      player_id: "p2",
      reason: "slurs in chat",
    });
  });
});

describe("the remove-and-block copy", () => {
  it("says the limit honestly: a guest with another name gets past it", () => {
    expect(BLOCK_HINT).toMatch(/guest/);
    expect(BLOCK_HINT).toMatch(/different name/);
    expect(BLOCK_HINT).toMatch(/new invite/);
  });
});

describe("the Reports tab", () => {
  it("parses #/admin/reports and nothing under it", () => {
    expect(parseHash("#/admin/reports")).toEqual({ name: "adminViews", view: { view: "reports" } });
    expect(parseHash("#/admin/reports/extra")).toEqual({ name: "adminLogin" });
  });
  it("has the hash it parses from and sits under its own tab", () => {
    expect(adminViewHash({ view: "reports" })).toBe(REPORTS_HASH);
    expect(REPORTS_HASH).toBe("#/admin/reports");
    expect(tabOf({ view: "reports" })).toBe("reports");
  });
});
