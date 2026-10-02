import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { isAdmin, loadAdminStatus, needsAdminCheck, resetAdminChecksForTest } from "./admin";
import { buildMenuSections } from "./contextMenu.logic";
import { canSwapSeats, gameWSURL } from "./gameURL";
import type { CardView, GameView, PlayerView } from "./protocol";
import { currentSession, setSession, type Session } from "./session";
import { canManageTable, canSpawn } from "./tableSettings";

// admin.test.ts covers ADR 0110 §3 item 4 on the client: who the UI
// treats as an admin. The shared token's session is one; so is a
// signed-in person whose Discord ID is on the server's allowlist, which
// the client learns from GET /me's `admin` and keeps on the session.
// The owner's requirement of 2026-10-02 is the last block: an admin
// seated at a table as themselves sees the admin UI there.

const GAME = "11111111-1111-1111-1111-111111111111";
const OTHER_GAME = "22222222-2222-2222-2222-222222222222";
const MY_SEAT = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
const THEIR_SEAT = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb";
const USER = "cccccccc-cccc-cccc-cccc-cccccccccccc";

function session(
  over: Partial<Session> & {
    role?: Session["principal"]["role"];
    user_id?: string;
  } = {},
): Session {
  const { role = "player", user_id, ...rest } = over;
  return {
    token: `tok-${Math.random()}`,
    expiresAt: "2099-01-01T00:00:00Z",
    principal: { role, user_id, issued_at: "", expires_at: "" },
    ...rest,
  } as Session;
}

function seatedAdmin(admin = true): Session {
  return session({
    role: "player",
    user_id: USER,
    gameID: GAME,
    playerID: MY_SEAT,
    admin,
  });
}

describe("isAdmin", () => {
  it("is true for the shared token and for an allowlisted person", () => {
    expect(isAdmin(session({ role: "admin" }))).toBe(true);
    expect(isAdmin(session({ role: "identified", user_id: USER, admin: true }))).toBe(true);
    expect(isAdmin(seatedAdmin())).toBe(true);
    expect(isAdmin(session({ role: "spectator", user_id: USER, admin: true }))).toBe(true);
  });

  it("is false for everyone else", () => {
    expect(isAdmin(null)).toBe(false);
    expect(isAdmin(undefined)).toBe(false);
    expect(isAdmin(seatedAdmin(false))).toBe(false);
    // Not asked yet.
    expect(isAdmin(session({ role: "player", user_id: USER }))).toBe(false);
    // No user: the server never makes such a session admin (a reclaim
    // ticket carries a seat's Discord ID and no user), so neither does
    // a stray flag.
    expect(isAdmin(session({ role: "player", admin: true }))).toBe(false);
  });
});

describe("loadAdminStatus", () => {
  beforeEach(() => resetAdminChecksForTest());
  afterEach(() => {
    vi.unstubAllGlobals();
    setSession(null);
  });

  function stubMe(body: unknown, status = 200): ReturnType<typeof vi.fn> {
    const fn = vi.fn(async () => new Response(JSON.stringify(body), { status }));
    vi.stubGlobal("fetch", fn);
    return fn;
  }

  it("asks /me once per session and records the answer", async () => {
    const s = session({ role: "player", user_id: USER, gameID: GAME, playerID: MY_SEAT });
    setSession(s);
    const fetchMock = stubMe({ role: "player", user_id: USER, admin: true });
    expect(needsAdminCheck(s)).toBe(true);
    await loadAdminStatus(s);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock.mock.calls[0][0]).toBe("/me");
    expect(currentSession()?.admin).toBe(true);
    expect(isAdmin(currentSession())).toBe(true);
    expect(needsAdminCheck(currentSession())).toBe(false);
  });

  it("records false for a person not on the list", async () => {
    const s = session({ role: "identified", user_id: USER, admin: true });
    setSession(s);
    stubMe({ role: "identified", user_id: USER, admin: false });
    await loadAdminStatus(s);
    expect(currentSession()?.admin).toBe(false);
    expect(isAdmin(currentSession())).toBe(false);
  });

  it("never asks for the token or a session with no user", () => {
    expect(needsAdminCheck(session({ role: "admin" }))).toBe(false);
    expect(needsAdminCheck(session({ role: "player" }))).toBe(false);
    expect(needsAdminCheck(null)).toBe(false);
  });

  it("leaves a session that was replaced meanwhile alone", async () => {
    const asked = session({ role: "identified", user_id: USER });
    const replaced = session({ role: "player", user_id: USER });
    setSession(asked);
    stubMe({ admin: true });
    const pending = loadAdminStatus(asked);
    setSession(replaced);
    await pending;
    expect(currentSession()?.token).toBe(replaced.token);
    expect(currentSession()?.admin).toBeUndefined();
  });

  it("keeps the session as it was when /me fails", async () => {
    const s = session({ role: "identified", user_id: USER, admin: true });
    setSession(s);
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("offline");
      }),
    );
    await loadAdminStatus(s);
    expect(currentSession()?.admin).toBe(true);
  });
});

// --- the owner's requirement: an admin seated as themselves ----------

function card(instance_id: string, owner: string): CardView {
  return { instance_id, name: instance_id, owner, controller: owner } as CardView;
}

function zone(kind: string, owner: string | undefined, cards: CardView[] = []) {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, extra: Partial<PlayerView> = {}): PlayerView {
  return {
    id,
    name: id,
    seat: 0,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    ...extra,
  } as PlayerView;
}

function tableView(battlefield: CardView[]): GameView {
  return {
    id: GAME,
    state: "active",
    seats: [seat(MY_SEAT), seat(THEIR_SEAT, { is_host: true })],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    turn: {
      seq: 1,
      number: 1,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
  } as unknown as GameView;
}

describe("a seated admin sees the admin UI", () => {
  const theirCard = card("their-creature", THEIR_SEAT);
  const view = tableView([theirCard]);

  it("gets the admin context menu on another seat's card", () => {
    const admin = seatedAdmin();
    expect(buildMenuSections(view, theirCard, MY_SEAT, isAdmin(admin)).length).toBeGreaterThan(0);
    // The same seat, not on the list: nothing to override.
    const player = seatedAdmin(false);
    expect(buildMenuSections(view, theirCard, MY_SEAT, isAdmin(player))).toEqual([]);
  });

  it("manages the table and may spawn, though it is not the host", () => {
    const mySeat = seat(MY_SEAT);
    expect(canManageTable("player", mySeat, isAdmin(seatedAdmin()))).toBe(true);
    expect(canManageTable("player", mySeat, isAdmin(seatedAdmin(false)))).toBe(false);
    const settings = { allow_spawn: true } as Parameters<typeof canSpawn>[2];
    expect(canSpawn("player", mySeat, settings, isAdmin(seatedAdmin()))).toBe(true);
    expect(canSpawn("player", mySeat, settings, isAdmin(seatedAdmin(false)))).toBe(false);
    // The token keeps working through the role default.
    expect(canManageTable("admin", null)).toBe(true);
  });

  it("plays as their own seat by default, and may swap seats like the token", () => {
    const admin = seatedAdmin();
    const url = new URL(gameWSURL({ baseURL: "wss://h/ws", gameID: GAME, session: admin }));
    expect(url.searchParams.get("player")).toBe(MY_SEAT);
    expect(canSwapSeats(admin)).toBe(true);
    const swapped = new URL(
      gameWSURL({ baseURL: "wss://h/ws", gameID: GAME, session: admin, seatOverride: THEIR_SEAT }),
    );
    expect(swapped.searchParams.get("player")).toBe(THEIR_SEAT);
    expect(canSwapSeats(seatedAdmin(false))).toBe(false);
  });

  it("opens another table without claiming a seat there", () => {
    const url = new URL(
      gameWSURL({ baseURL: "wss://h/ws", gameID: OTHER_GAME, session: seatedAdmin() }),
    );
    expect(url.searchParams.get("game")).toBe(OTHER_GAME);
    expect(url.searchParams.get("player")).toBeNull();
  });
});
