// community.test.ts — GET /community, read once, and every failure is
// "no invite, no gate" (ADR 0156 §6): the client works before the server
// route exists, and every Discord placement hides.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";
import {
  _resetCommunityForTests,
  community,
  loadCommunity,
  NO_COMMUNITY,
  parseCommunity,
} from "./community";

const INVITE = "https://discord.gg/abc123";

describe("parseCommunity", () => {
  it("reads the invite and the gate", () => {
    expect(parseCommunity({ invite_url: INVITE, gate: true })).toEqual({
      invite_url: INVITE,
      gate: true,
      voice: false,
    });
  });

  it("reads whether voice rooms are offered, and only a literal true (ADR 0159 §7)", () => {
    expect(parseCommunity({ invite_url: INVITE, gate: false, voice: true }).voice).toBe(true);
    expect(parseCommunity({ invite_url: INVITE, gate: false, voice: "yes" }).voice).toBe(false);
    expect(parseCommunity({ invite_url: INVITE, gate: false }).voice).toBe(false);
  });

  it("takes only an absolute https URL as the invite", () => {
    for (const bad of ["", "discord.gg/abc", "http://discord.gg/abc", "javascript:alert(1)", 7]) {
      expect(parseCommunity({ invite_url: bad, gate: false }).invite_url).toBeNull();
    }
  });

  it("is off for anything but a literal true gate, and for a body that is not an object", () => {
    expect(parseCommunity({ invite_url: INVITE, gate: "yes" }).gate).toBe(false);
    expect(parseCommunity(null)).toEqual(NO_COMMUNITY);
    expect(parseCommunity("x")).toEqual(NO_COMMUNITY);
  });
});

describe("loadCommunity", () => {
  beforeEach(() => _resetCommunityForTests());
  afterEach(() => vi.unstubAllGlobals());

  it("starts with no invite and no gate", () => {
    expect(get(community)).toEqual(NO_COMMUNITY);
  });

  it("fills the store from the server, once", async () => {
    const f = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ invite_url: INVITE, gate: true }),
    });
    vi.stubGlobal("fetch", f);
    await Promise.all([loadCommunity(), loadCommunity()]);
    await loadCommunity();
    expect(get(community)).toEqual({ invite_url: INVITE, gate: true, voice: false });
    expect(f).toHaveBeenCalledTimes(1);
    expect(f.mock.calls[0][0]).toBe("/community");
  });

  it("falls back on a 404, a server that has no route yet", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 404 }));
    expect(await loadCommunity()).toEqual(NO_COMMUNITY);
    expect(get(community)).toEqual(NO_COMMUNITY);
  });

  it("falls back on a network error and on a body that is not JSON", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("offline")));
    expect(await loadCommunity()).toEqual(NO_COMMUNITY);

    _resetCommunityForTests();
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => {
          throw new SyntaxError("<!doctype html>");
        },
      }),
    );
    expect(await loadCommunity()).toEqual(NO_COMMUNITY);
  });
});
