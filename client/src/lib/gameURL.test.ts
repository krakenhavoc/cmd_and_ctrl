import { describe, it, expect } from "vitest";
import { gameWSURL, canSwapSeats } from "./gameURL";
import type { Session } from "./session";

const BASE = "wss://host/ws";
const GAME = "11111111-1111-1111-1111-111111111111";
const SEAT_A = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
const SEAT_B = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb";

function session(over: Partial<Session> & { role?: Session["principal"]["role"] }): Session {
  const { role = "player", ...rest } = over;
  return {
    token: "tok",
    expiresAt: "2099-01-01T00:00:00Z",
    principal: { role, issued_at: "", expires_at: "" },
    ...rest,
  } as Session;
}

function seatOf(url: string): string | null {
  return new URL(url).searchParams.get("player");
}

describe("canSwapSeats", () => {
  it("is admin-only", () => {
    expect(canSwapSeats(session({ role: "admin" }))).toBe(true);
    expect(canSwapSeats(session({ role: "player" }))).toBe(false);
    expect(canSwapSeats(session({ role: "spectator" }))).toBe(false);
    expect(canSwapSeats(null)).toBe(false);
  });
});

describe("gameWSURL", () => {
  it("carries the game and token", () => {
    const u = new URL(gameWSURL({ baseURL: BASE, gameID: GAME, session: session({}) }));
    expect(u.searchParams.get("game")).toBe(GAME);
    expect(u.searchParams.get("token")).toBe("tok");
  });

  it("pins a player session to its own seat", () => {
    const s = session({ role: "player", playerID: SEAT_A, gameID: GAME });
    expect(seatOf(gameWSURL({ baseURL: BASE, gameID: GAME, session: s }))).toBe(SEAT_A);
  });

  // A seat bound to a DIFFERENT game must not leak across; the route
  // can mount for game B while the stored session is still game A's.
  it("ignores a bound seat from another game", () => {
    const s = session({ role: "player", playerID: SEAT_A, gameID: "other" });
    expect(seatOf(gameWSURL({ baseURL: BASE, gameID: GAME, session: s }))).toBeNull();
  });

  it("omits the seat for an admin with no override — the spectator view", () => {
    const s = session({ role: "admin" });
    expect(seatOf(gameWSURL({ baseURL: BASE, gameID: GAME, session: s }))).toBeNull();
    expect(
      seatOf(gameWSURL({ baseURL: BASE, gameID: GAME, session: s, seatOverride: null })),
    ).toBeNull();
  });

  it("lets an admin override the seat", () => {
    const s = session({ role: "admin" });
    expect(
      seatOf(gameWSURL({ baseURL: BASE, gameID: GAME, session: s, seatOverride: SEAT_B })),
    ).toBe(SEAT_B);
  });

  // The important negative: a non-admin cannot talk itself onto
  // another seat by setting the override. The server independently
  // refuses too, but the client must not render as though it worked.
  it("ignores an override on a player session", () => {
    const s = session({ role: "player", playerID: SEAT_A, gameID: GAME });
    expect(
      seatOf(gameWSURL({ baseURL: BASE, gameID: GAME, session: s, seatOverride: SEAT_B })),
    ).toBe(SEAT_A);
  });

  it("ignores an override on a spectator session", () => {
    const s = session({ role: "spectator" });
    expect(
      seatOf(gameWSURL({ baseURL: BASE, gameID: GAME, session: s, seatOverride: SEAT_B })),
    ).toBeNull();
  });

  it("handles a missing session", () => {
    const u = new URL(gameWSURL({ baseURL: BASE, gameID: GAME, session: null }));
    expect(u.searchParams.get("token")).toBeNull();
    expect(u.searchParams.get("player")).toBeNull();
  });
});
