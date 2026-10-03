import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  adminChipFor,
  adminChipLabel,
  adminModeVerdict,
  afterAdminModeChanged,
  armAdminLapse,
  clockTime,
  isAdmin,
  isAdminAllowed,
  LAPSE_RECHECK_MIN_MS,
  loadAdminStatus,
  needsAdminCheck,
  onVisibleAgain,
  ownBindingFor,
  resetAdminChecksForTest,
  switchAdminMode,
} from "./admin";
import { buildMenuSections } from "./contextMenu.logic";
import { canSwapSeats, gameWSURL } from "./gameURL";
import type { CardView, GameView, PlayerView } from "./protocol";
import { currentSession, LobbyApiError, setSession, type Session } from "./session";
import { canManageTable, canSpawn } from "./tableSettings";
import { canCreateTables } from "./tableSetup";

// admin.test.ts covers who the client treats as an admin: ADR 0110 §3
// item 4, as ADR 0112 §2 amends it. The shared token's session is one.
// So is a signed-in person whose Discord ID is on the server's
// allowlist, but only in admin mode, which they switch on with the
// Admin chip and which lapses 12 hours later (owner answer 1). The
// client learns all of it from GET /me (`admin_allowed`, `admin_mode`,
// `admin_mode_ends_at`) and keeps it on the session. The owner's
// requirement of 2026-10-02 is the last block: an admin seated at a
// table as themselves sees the admin UI there, in admin mode.

const GAME = "11111111-1111-1111-1111-111111111111";
const OTHER_GAME = "22222222-2222-2222-2222-222222222222";
const MY_SEAT = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
const THEIR_SEAT = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb";
const USER = "cccccccc-cccc-cccc-cccc-cccccccccccc";
const HOUR = 3_600_000;

function session(
  over: Partial<Session> & {
    role?: Session["principal"]["role"];
    user_id?: string;
    game_id?: string;
  } = {},
): Session {
  const { role = "player", user_id, game_id, ...rest } = over;
  return {
    token: `tok-${Math.random()}`,
    expiresAt: "2099-01-01T00:00:00Z",
    principal: { role, user_id, game_id, issued_at: "", expires_at: "" },
    ...rest,
  } as Session;
}

// inAdminMode is what /me says about an allowlisted person with admin
// mode on until `ends`.
function inAdminMode(ends = Date.now() + 12 * HOUR): Partial<Session> {
  return { admin: true, admin_allowed: true, admin_mode: true, admin_mode_ends_at: ends };
}

// inPlayerMode is the same person with admin mode off.
const inPlayerMode: Partial<Session> = { admin: false, admin_allowed: true, admin_mode: false };

function seatedAdmin(admin = true): Session {
  return session({
    role: "player",
    user_id: USER,
    gameID: GAME,
    playerID: MY_SEAT,
    ...(admin ? inAdminMode() : inPlayerMode),
  });
}

describe("isAdmin", () => {
  it("is true for the shared token and for an allowlisted person in admin mode", () => {
    expect(isAdmin(session({ role: "admin" }))).toBe(true);
    expect(isAdmin(session({ role: "identified", user_id: USER, ...inAdminMode() }))).toBe(true);
    expect(isAdmin(seatedAdmin())).toBe(true);
    expect(isAdmin(session({ role: "spectator", user_id: USER, ...inAdminMode() }))).toBe(true);
  });

  it("is false for an allowlisted person in player mode (ADR 0112 §2)", () => {
    expect(isAdmin(seatedAdmin(false))).toBe(false);
    expect(isAdmin(session({ role: "identified", user_id: USER, ...inPlayerMode }))).toBe(false);
  });

  it("is false once admin mode has lapsed, before anyone says so", () => {
    const ends = Date.now() + HOUR;
    const s = session({ role: "player", user_id: USER, ...inAdminMode(ends) });
    expect(isAdmin(s, ends - 1)).toBe(true);
    expect(isAdmin(s, ends)).toBe(false);
    expect(isAdmin(s, ends + HOUR)).toBe(false);
  });

  it("reads a stale or half-filled answer as no", () => {
    // A session stored before ADR 0112 carries only `admin`.
    expect(isAdmin(session({ role: "player", user_id: USER, admin: true }))).toBe(false);
    // Admin mode with no end time is not something the server sends.
    expect(
      isAdmin(
        session({
          role: "player",
          user_id: USER,
          admin_allowed: true,
          admin_mode: true,
          admin: true,
        }),
      ),
    ).toBe(false);
    // A mode on a person who is not allowed.
    expect(
      isAdmin(
        session({
          role: "player",
          user_id: USER,
          admin_mode: true,
          admin_mode_ends_at: Date.now() + HOUR,
        }),
      ),
    ).toBe(false);
  });

  it("is false for everyone else", () => {
    expect(isAdmin(null)).toBe(false);
    expect(isAdmin(undefined)).toBe(false);
    // Not asked yet.
    expect(isAdmin(session({ role: "player", user_id: USER }))).toBe(false);
    // No user: the server never makes such a session admin (a reclaim
    // ticket carries a seat's Discord ID and no user), so neither does
    // a stray flag.
    expect(isAdmin(session({ role: "player", ...inAdminMode() }))).toBe(false);
  });
});

describe("isAdminAllowed", () => {
  it("is the allowlisted person in either mode, and nobody else", () => {
    expect(isAdminAllowed(seatedAdmin())).toBe(true);
    expect(isAdminAllowed(seatedAdmin(false))).toBe(true);
    expect(isAdminAllowed(session({ role: "admin" }))).toBe(false);
    expect(isAdminAllowed(session({ role: "player", user_id: USER }))).toBe(false);
    expect(isAdminAllowed(session({ role: "player", admin_allowed: true }))).toBe(false);
    expect(isAdminAllowed(null)).toBe(false);
  });
});

describe("the chip (ADR 0112 §2 item 9)", () => {
  it("is a switch for an allowlisted person, showing the mode and when it ends", () => {
    const ends = new Date(2026, 9, 2, 23, 40).getTime();
    const now = ends - HOUR;
    expect(adminChipFor(seatedAdmin(false), now)).toEqual({
      kind: "switch",
      on: false,
      endsAt: null,
    });
    const on = session({ role: "identified", user_id: USER, ...inAdminMode(ends) });
    expect(adminChipFor(on, now)).toEqual({ kind: "switch", on: true, endsAt: ends });
    expect(adminChipLabel(true, ends)).toBe("Admin · until 23:40");
    expect(adminChipLabel(false, null)).toBe("Player");
    // Lapsed: "Player" again.
    expect(adminChipFor(on, ends + 1)).toEqual({ kind: "switch", on: false, endsAt: null });
  });

  it("is the static badge for the token, and nothing for anyone else", () => {
    expect(adminChipFor(session({ role: "admin" }))).toEqual({ kind: "token" });
    expect(adminChipFor(session({ role: "player", user_id: USER, admin: false }))).toBeNull();
    expect(adminChipFor(session({ role: "spectator" }))).toBeNull();
    expect(adminChipFor(null)).toBeNull();
  });

  it("formats the end time as the local HH:MM", () => {
    expect(clockTime(new Date(2026, 0, 1, 7, 5).getTime())).toBe("07:05");
    expect(clockTime(Number.NaN)).toBe("");
  });
});

describe("every isAdmin reader follows the mode", () => {
  // The grep of client/src for isAdmin( (ADR 0112 §2, "The client
  // follows admin"): gameURL's seat swap, tableSetup's create form, and
  // the boolean Lobby and Game hand to the context menu and the table
  // gates. Each must treat player mode as an ordinary person.
  it("seat swap is admin mode only", () => {
    expect(canSwapSeats(seatedAdmin())).toBe(true);
    expect(canSwapSeats(seatedAdmin(false))).toBe(false);
  });

  it("an allowlisted person may create tables in either mode, as any signed-in person may", () => {
    expect(canCreateTables(seatedAdmin())).toBe(true);
    expect(canCreateTables(seatedAdmin(false))).toBe(true);
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

  it("asks /me once per session and records the whole answer", async () => {
    const s = session({ role: "player", user_id: USER, gameID: GAME, playerID: MY_SEAT });
    setSession(s);
    const ends = Date.now() + 12 * HOUR;
    const fetchMock = stubMe({
      role: "player",
      user_id: USER,
      admin: true,
      admin_allowed: true,
      admin_mode: true,
      admin_mode_ends_at: ends,
    });
    expect(needsAdminCheck(s)).toBe(true);
    expect(await loadAdminStatus(s)).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock.mock.calls[0][0]).toBe("/me");
    const cur = currentSession();
    expect(cur?.admin_allowed).toBe(true);
    expect(cur?.admin_mode).toBe(true);
    expect(cur?.admin_mode_ends_at).toBe(ends);
    expect(isAdmin(cur)).toBe(true);
    expect(needsAdminCheck(cur)).toBe(false);
  });

  it("records player mode for an allowlisted person with the mode off", async () => {
    const s = session({ role: "identified", user_id: USER, ...inAdminMode() });
    setSession(s);
    stubMe({ admin: false, admin_allowed: true, admin_mode: false });
    await loadAdminStatus(s);
    const cur = currentSession();
    expect(cur?.admin_allowed).toBe(true);
    expect(cur?.admin_mode).toBe(false);
    expect(cur?.admin_mode_ends_at).toBeUndefined();
    expect(isAdmin(cur)).toBe(false);
  });

  it("records false for a person not on the list", async () => {
    const s = session({ role: "identified", user_id: USER, ...inAdminMode() });
    setSession(s);
    stubMe({ role: "identified", user_id: USER, admin: false });
    await loadAdminStatus(s);
    expect(currentSession()?.admin_allowed).toBe(false);
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
    stubMe({ admin: true, admin_allowed: true });
    const pending = loadAdminStatus(asked);
    setSession(replaced);
    await pending;
    expect(currentSession()?.token).toBe(replaced.token);
    expect(currentSession()?.admin_allowed).toBeUndefined();
  });

  it("keeps the session as it was when /me fails, and says so", async () => {
    const s = session({ role: "identified", user_id: USER, ...inAdminMode() });
    setSession(s);
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("offline");
      }),
    );
    expect(await loadAdminStatus(s)).toBe(false);
    expect(isAdmin(currentSession())).toBe(true);
  });
});

describe("switchAdminMode (PUT /me/admin-mode)", () => {
  beforeEach(() => resetAdminChecksForTest());
  afterEach(() => {
    vi.unstubAllGlobals();
    setSession(null);
  });

  it("sends {on} and takes the answer onto the session", async () => {
    setSession(session({ role: "player", user_id: USER, ...inPlayerMode }));
    const ends = Date.now() + 12 * HOUR;
    const fetchMock = vi.fn(
      async () =>
        new Response(JSON.stringify({ admin: true, admin_mode: true, admin_mode_ends_at: ends }), {
          status: 200,
        }),
    );
    vi.stubGlobal("fetch", fetchMock);
    await switchAdminMode(true);
    const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe("/me/admin-mode");
    expect(init.method).toBe("PUT");
    expect(JSON.parse(String(init.body))).toEqual({ on: true });
    expect(isAdmin(currentSession())).toBe(true);
    expect(currentSession()?.admin_mode_ends_at).toBe(ends);

    fetchMock.mockImplementationOnce(
      async () =>
        new Response(JSON.stringify({ admin: false, admin_mode: false }), { status: 200 }),
    );
    await switchAdminMode(false);
    expect(isAdmin(currentSession())).toBe(false);
    expect(currentSession()?.admin_allowed).toBe(true);
    expect(currentSession()?.admin_mode_ends_at).toBeUndefined();
  });

  it("leaves the session as it was on a refusal, and throws the server's message", async () => {
    const before = session({ role: "player", user_id: USER, ...inPlayerMode });
    setSession(before);
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () =>
          new Response(JSON.stringify({ error: "lobby: not an admin" }), {
            status: 403,
            statusText: "Forbidden",
          }),
      ),
    );
    const err = await switchAdminMode(true).catch((e: unknown) => e);
    expect(err).toBeInstanceOf(LobbyApiError);
    expect((err as LobbyApiError).status).toBe(403);
    expect((err as Error).message).toContain("not an admin");
    expect(currentSession()).toEqual(before);
  });
});

describe("asking /me again (ADR 0112 §2 item 9)", () => {
  beforeEach(() => {
    resetAdminChecksForTest();
    vi.useFakeTimers();
  });
  afterEach(() => {
    armAdminLapse(null);
    setSession(null);
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  function stubMeSequence(...bodies: unknown[]): ReturnType<typeof vi.fn> {
    let i = 0;
    const fn = vi.fn(async () => {
      const body = bodies[Math.min(i, bodies.length - 1)];
      i += 1;
      return new Response(JSON.stringify(body), { status: 200 });
    });
    vi.stubGlobal("fetch", fn);
    return fn;
  }

  it("ends admin mode when its end time passes, then asks /me", async () => {
    const ends = Date.now() + 2 * HOUR;
    setSession(session({ role: "player", user_id: USER, ...inAdminMode(ends) }));
    const fetchMock = stubMeSequence({ admin: false, admin_allowed: true, admin_mode: false });
    armAdminLapse(currentSession());
    await vi.advanceTimersByTimeAsync(2 * HOUR - 1000);
    expect(fetchMock).not.toHaveBeenCalled();
    expect(currentSession()?.admin_mode).toBe(true);
    await vi.advanceTimersByTimeAsync(1000);
    expect(currentSession()?.admin_mode).toBe(false);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock.mock.calls[0][0]).toBe("/me");
  });

  it("asks no more than once a minute when the server's clock is behind", async () => {
    // The server still says "on, ending at a time already past here":
    // a client whose clock runs ahead must not re-ask in a tight loop.
    const ends = Date.now() + 1000;
    setSession(session({ role: "player", user_id: USER, ...inAdminMode(ends) }));
    const stillOn = {
      admin: true,
      admin_allowed: true,
      admin_mode: true,
      admin_mode_ends_at: ends,
    };
    const fetchMock = stubMeSequence(stillOn);
    // App.svelte re-arms on every session change; so does this.
    const unsubscribe = (await import("./session")).session.subscribe((s) => armAdminLapse(s));
    await vi.advanceTimersByTimeAsync(1000);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(LAPSE_RECHECK_MIN_MS - 1000);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(5 * LAPSE_RECHECK_MIN_MS);
    expect(fetchMock.mock.calls.length).toBeLessThanOrEqual(7);
    expect(isAdmin(currentSession())).toBe(false);
    unsubscribe();
  });

  it("asks again when the tab becomes visible, for an allowlisted person only", async () => {
    setSession(session({ role: "player", user_id: USER, ...inPlayerMode }));
    const ends = Date.now() + 12 * HOUR;
    const fetchMock = stubMeSequence({
      admin: true,
      admin_allowed: true,
      admin_mode: true,
      admin_mode_ends_at: ends,
    });
    onVisibleAgain();
    await vi.advanceTimersByTimeAsync(0);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(isAdmin(currentSession())).toBe(true);

    setSession(session({ role: "player", user_id: USER, admin: false, admin_allowed: false }));
    onVisibleAgain();
    setSession(session({ role: "admin" }));
    onVisibleAgain();
    await vi.advanceTimersByTimeAsync(0);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});

describe("after a 4001 (ADR 0112 §2 item 5)", () => {
  beforeEach(() => resetAdminChecksForTest());
  afterEach(() => {
    vi.unstubAllGlobals();
    setSession(null);
  });

  it("knows which bindings need no admin rights", () => {
    const own = session({ role: "player", user_id: USER, gameID: GAME, playerID: MY_SEAT });
    expect(ownBindingFor(own, GAME)).toBe(true);
    expect(ownBindingFor(own, OTHER_GAME)).toBe(false);
    const watching = session({ role: "spectator", user_id: USER, gameID: GAME });
    expect(ownBindingFor(watching, GAME)).toBe(true);
    expect(ownBindingFor(watching, OTHER_GAME)).toBe(false);
    expect(ownBindingFor(session({ role: "identified", user_id: USER }), GAME)).toBe(false);
    expect(ownBindingFor(null, GAME)).toBe(false);
  });

  it("reconnects an admin anywhere, a player at their own table, and sends the rest away", () => {
    const elsewhere = session({ role: "player", user_id: USER, gameID: OTHER_GAME });
    expect(adminModeVerdict({ ...elsewhere, ...inAdminMode() }, GAME)).toBe("reconnect");
    expect(adminModeVerdict({ ...elsewhere, ...inPlayerMode }, GAME)).toBe("leave");
    expect(adminModeVerdict(seatedAdmin(false), GAME)).toBe("reconnect");
    const identified = session({ role: "identified", user_id: USER, ...inPlayerMode });
    expect(adminModeVerdict(identified, GAME)).toBe("leave");
    expect(adminModeVerdict(session({ role: "admin" }), GAME)).toBe("reconnect");
  });

  it("asks /me before deciding, and says retry when /me does not answer", async () => {
    setSession(session({ role: "identified", user_id: USER, ...inAdminMode() }));
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response(JSON.stringify(inPlayerMode), { status: 200 })),
    );
    expect(await afterAdminModeChanged(GAME)).toBe("leave");

    setSession(session({ role: "identified", user_id: USER, ...inAdminMode() }));
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("offline");
      }),
    );
    expect(await afterAdminModeChanged(GAME)).toBe("retry");
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

describe("a seated admin sees the admin UI in admin mode", () => {
  const theirCard = card("their-creature", THEIR_SEAT);
  const view = tableView([theirCard]);

  it("gets the admin context menu on another seat's card", () => {
    const admin = seatedAdmin();
    expect(buildMenuSections(view, theirCard, MY_SEAT, isAdmin(admin)).length).toBeGreaterThan(0);
    // The same seat in player mode: nothing to override.
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
    // In player mode the override is ignored: their own seat.
    const asPlayer = new URL(
      gameWSURL({
        baseURL: "wss://h/ws",
        gameID: GAME,
        session: seatedAdmin(false),
        seatOverride: THEIR_SEAT,
      }),
    );
    expect(asPlayer.searchParams.get("player")).toBe(MY_SEAT);
  });

  it("opens another table without claiming a seat there", () => {
    const url = new URL(
      gameWSURL({ baseURL: "wss://h/ws", gameID: OTHER_GAME, session: seatedAdmin() }),
    );
    expect(url.searchParams.get("game")).toBe(OTHER_GAME);
    expect(url.searchParams.get("player")).toBeNull();
  });
});
