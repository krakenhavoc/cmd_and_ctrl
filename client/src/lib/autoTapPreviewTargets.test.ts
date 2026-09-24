import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { fetchAutoTapPreview } from "./api";
import { setSession, type Session } from "./session";

// #1405: the activation preview prices a target-keyed cost (Dragonfire
// Blade's "{1} less for each color of the creature it targets") at the
// target's price when the request names the target. These pin the
// client half: fetchAutoTapPreview sends the targets it is given, in
// the `<kind>:<uuid>` form the server parses, and only on the ability
// branch.

const session: Session = {
  token: "tok",
  expiresAt: new Date(Date.now() + 3_600_000).toISOString(),
  principal: {
    role: "player",
    game_id: "g1",
    player_id: "p1",
    issued_at: new Date().toISOString(),
    expires_at: new Date(Date.now() + 3_600_000).toISOString(),
  },
};

let urls: string[] = [];

function queryOf(url: string): URLSearchParams {
  return new URL(url, "http://localhost").searchParams;
}

describe("fetchAutoTapPreview targets", () => {
  beforeEach(() => {
    urls = [];
    setSession(session);
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: string) => {
        urls.push(input);
        return {
          ok: true,
          status: 200,
          statusText: "stub",
          json: async () => ({ ok: true, cost: "{4}" }),
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

  it("sends the activation's targets as kind:id", async () => {
    await fetchAutoTapPreview("g1", "blade", {
      abilityIndex: 0,
      targets: [
        { kind: "card", id: "vivi" },
        { kind: "player", id: "bob" },
      ],
    });
    const q = queryOf(urls[0]);
    expect(q.get("ability")).toBe("0");
    expect(q.get("targets")).toBe("card:vivi,player:bob");
  });

  it("omits targets it cannot name and sends none when it has none", async () => {
    await fetchAutoTapPreview("g1", "blade", {
      abilityIndex: 0,
      targets: [{ kind: "self" }, { kind: "none" }],
    });
    await fetchAutoTapPreview("g1", "blade", { abilityIndex: 0 });
    expect(queryOf(urls[0]).has("targets")).toBe(false);
    expect(queryOf(urls[1]).has("targets")).toBe(false);
  });

  it("does not send targets on the cast branch", async () => {
    await fetchAutoTapPreview("g1", "bolt", { targets: [{ kind: "card", id: "x" }] });
    expect(queryOf(urls[0]).has("targets")).toBe(false);
  });
});
