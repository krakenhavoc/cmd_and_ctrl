// blockList.test.ts — the client half of #3113's logic: who is offered the
// block list, the requests it and the spectator controls send, and the
// copy that has to stay honest.

import { afterEach, describe, expect, it, vi } from "vitest";

import { fetchBlocks, fetchSpectators, removeSpectator, reportPlayer, unblock } from "./api";
import { LobbyApiError } from "./session";
import {
  BLOCK_LIST_HINT,
  SPECTATORS_HINT,
  blockErrorMessage,
  blockKindLabel,
  canManageBlocks,
  spectatorErrorMessage,
  spectatorTargets,
} from "./moderation";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("canManageBlocks mirrors the server's authority split", () => {
  const lobby = { state: "lobby" };
  const started = { state: "active" };
  const ended = { state: "ended" };

  it("gives the host and the creator the list on an unstarted table", () => {
    expect(canManageBlocks(lobby, { isAdmin: false, isHost: true })).toBe(true);
    expect(canManageBlocks({ ...lobby, is_creator: true }, { isAdmin: false, isHost: false })).toBe(
      true,
    );
  });

  it("gives nobody else the list: another seat, a spectator, a stranger", () => {
    expect(canManageBlocks(lobby, { isAdmin: false, isHost: false })).toBe(false);
    expect(
      canManageBlocks({ ...lobby, is_creator: false }, { isAdmin: false, isHost: false }),
    ).toBe(false);
  });

  it("makes a table that has started, or ended, the admin's alone", () => {
    for (const g of [started, ended]) {
      expect(canManageBlocks(g, { isAdmin: false, isHost: true })).toBe(false);
      expect(canManageBlocks({ ...g, is_creator: true }, { isAdmin: false, isHost: false })).toBe(
        false,
      );
      expect(canManageBlocks(g, { isAdmin: true, isHost: false })).toBe(true);
    }
    expect(canManageBlocks(lobby, { isAdmin: true, isHost: false })).toBe(true);
  });
});

describe("the requests", () => {
  function stubFetch(body: unknown): { calls: { url: string; init: RequestInit }[] } {
    const calls: { url: string; init: RequestInit }[] = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init: RequestInit) => {
        calls.push({ url, init });
        return new Response(JSON.stringify(body), { status: 200 });
      }),
    );
    return { calls };
  }

  it("fetchBlocks reads the table's list", async () => {
    const entry = { id: "e1", label: "bob", kind: "guest", created_at: 1 };
    const { calls } = stubFetch({ blocks: [entry] });
    expect(await fetchBlocks("g1")).toEqual([entry]);
    expect(calls[0].url).toBe("/games/g1/blocks");
    expect(calls[0].init.method).toBe("GET");
  });

  it("unblock deletes the entry and returns the list as it now stands", async () => {
    const { calls } = stubFetch({ blocks: [] });
    expect(await unblock("g1", "e 1")).toEqual([]);
    expect(calls[0].url).toBe("/games/g1/blocks/e%201");
    expect(calls[0].init.method).toBe("DELETE");
  });

  it("fetchSpectators reads who is watching", async () => {
    const { calls } = stubFetch({ spectators: [{ id: "s1", name: "Eve", since: 1 }] });
    expect((await fetchSpectators("g1"))[0].name).toBe("Eve");
    expect(calls[0].url).toBe("/games/g1/spectators");
  });

  it("removeSpectator sends no block by default and ?block=true on request", async () => {
    const { calls } = stubFetch({ blocked: true, spectators: [] });
    await removeSpectator("g1", "s1");
    const out = await removeSpectator("g1", "s1", { block: true });
    expect(out.blocked).toBe(true);
    expect(calls.map((c) => c.url)).toEqual([
      "/games/g1/spectators/s1",
      "/games/g1/spectators/s1?block=true",
    ]);
    expect(calls.every((c) => c.init.method === "DELETE")).toBe(true);
  });

  it("reportPlayer names a spectator by spectator_id, never both ids", async () => {
    const { calls } = stubFetch({ id: "r1" });
    await reportPlayer("g1", "s1", "slurs", "spectator");
    await reportPlayer("g1", "p2", "slurs");
    expect(JSON.parse(calls[0].init.body as string)).toEqual({
      spectator_id: "s1",
      reason: "slurs",
    });
    expect(JSON.parse(calls[1].init.body as string)).toEqual({ player_id: "p2", reason: "slurs" });
  });
});

describe("report targets and messages", () => {
  it("marks each watching session as a spectator target", () => {
    expect(spectatorTargets([{ id: "s1", name: "Eve" }])).toEqual([
      { id: "s1", name: "Eve", kind: "spectator" },
    ]);
  });

  it("labels what a block holds against", () => {
    expect(blockKindLabel("account")).toBe("account");
    expect(blockKindLabel("guest")).toBe("guest name");
  });

  it("shows the server's sentence, or a plain one for a block that is already gone", () => {
    expect(blockErrorMessage(new LobbyApiError(404, "no such block at this table"))).toMatch(
      /already gone/,
    );
    expect(blockErrorMessage(new LobbyApiError(409, "the game has started"))).toBe(
      "the game has started",
    );
    expect(spectatorErrorMessage(new LobbyApiError(404, "x"))).toMatch(/isn't watching/);
    expect(spectatorErrorMessage(new LobbyApiError(403, "not yours"))).toBe("not yours");
  });
});

describe("the copy says the limits honestly", () => {
  it("names the guest-name limit on the list and on the spectator panel", () => {
    expect(BLOCK_LIST_HINT).toMatch(/guest block holds against the name only/);
    expect(SPECTATORS_HINT).toMatch(/guest with no name cannot be blocked/);
    expect(SPECTATORS_HINT).toMatch(/spectator link/);
  });
});
