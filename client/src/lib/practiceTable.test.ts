// @vitest-environment jsdom
//
// practiceTable.test.ts — the tutorial's practice table, client half
// (ADR 0076 §2.2, #1078). The point of the module is that the four
// forced settings and the swapped session come back on EVERY exit, so
// that is most of what is tested here: the Leave button (a route
// change), a closed tab (pagehide), and an exit that ran no code at
// all (a fresh page load finding the record).
//
// jsdom for localStorage, sessionStorage and a window to dispatch
// pagehide on. The router and the two API calls are mocked: the route
// is a plain store the test drives, and the network is a spy.

import { beforeEach, describe, expect, it, vi, type Mock } from "vitest";
import { get, writable, type Writable } from "svelte/store";
import type { Session } from "./session";
import type { Route } from "./router";

const SETTINGS_KEY = "cmdctrl.settings.v1";
const SESSION_KEY = "cmdctrl.session";

function session(token: string, extra: Partial<Session> = {}): Session {
  const expires = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token,
    expiresAt: expires,
    principal: { role: "player", issued_at: new Date().toISOString(), expires_at: expires },
    ...extra,
  };
}

const PRACTICE = session("practice-token", { gameID: "practice-game", playerID: "seat-0" });
const MINE = session("my-real-seat", { gameID: "real-game", playerID: "seat-9" });

// load imports the modules fresh — a new page, as far as their
// module-level state is concerned — over whatever localStorage holds.
// Each page gets its own route store and network spies (doMock, not a
// hoisted vi.mock, whose instances would outlive resetModules): two
// tabs do not share a route.
async function load() {
  vi.resetModules();
  vi.doMock("./router", () => ({
    route: writable<Route>({ name: "lobby" }),
    navigate: vi.fn(),
  }));
  vi.doMock("./api", () => ({
    createPracticeTable: vi.fn(),
    leavePracticeTable: vi.fn(async () => {}),
  }));
  const practice = await import("./practiceTable");
  const settingsMod = await import("./settings");
  const sessionMod = await import("./session");
  const router = await import("./router");
  const api = await import("./api");
  return {
    ...practice,
    settings: settingsMod.settings,
    updateSettings: settingsMod.updateSettings,
    currentSession: sessionMod.currentSession,
    setSession: sessionMod.setSession,
    route: router.route as unknown as Writable<Route>,
    navigate: router.navigate as unknown as Mock,
    createPracticeTable: api.createPracticeTable as unknown as Mock,
    leavePracticeTable: api.leavePracticeTable as unknown as Mock,
  };
}

// storedSettings reads what is on disk, which is what the next page
// load — and every real game afterwards — will see.
function storedSettings() {
  const s = JSON.parse(localStorage.getItem(SETTINGS_KEY) ?? "{}");
  return {
    strictMana: s.gameplay?.strictMana,
    passMode: s.gameplay?.passMode,
    tableLayout: s.display?.tableLayout,
    cardSize: s.display?.cardSize,
  };
}

const MY_SETTINGS = {
  strictMana: false,
  passMode: "careful",
  tableLayout: "row",
  cardSize: "large",
} as const;

// openPractice puts a player with non-default settings at a real
// table, then opens the tutorial.
async function openPractice() {
  const m = await load();
  m.setSession(MINE);
  m.updateSettings("gameplay", "strictMana", false);
  m.updateSettings("gameplay", "passMode", "careful");
  m.updateSettings("display", "tableLayout", "row");
  m.updateSettings("display", "cardSize", "large");
  m.installPracticeExits();
  m.route.set({ name: "practice" });
  m.createPracticeTable.mockResolvedValue(PRACTICE);
  const id = await m.startPractice();
  m.route.set({ name: "game", gameID: id });
  return m;
}

beforeEach(() => {
  // Every earlier test's module instance is still listening on this
  // window. End whatever they are running before wiping the storage
  // they would otherwise write back into mid-test.
  window.dispatchEvent(new Event("pagehide"));
  vi.resetModules();
  vi.clearAllMocks();
  localStorage.clear();
  sessionStorage.clear();
});

describe("opening the practice table", () => {
  it("records what it swaps, then forces the four settings and takes the practice seat", async () => {
    const m = await openPractice();
    expect(m.createPracticeTable).toHaveBeenCalledOnce();
    expect(get(m.practiceTable)).toEqual({ gameID: "practice-game" });
    expect(m.isPracticeGame("practice-game")).toBe(true);
    expect(m.currentSession()?.token).toBe("practice-token");
    expect(storedSettings()).toEqual(m.FORCED_SETTINGS);

    const rec = m.parseRecord(localStorage.getItem(m.RECORD_KEY));
    expect(rec?.gameID).toBe("practice-game");
    expect(rec?.saved).toEqual(MY_SETTINGS);
    expect(rec?.previous?.token).toBe("my-real-seat");
  });

  it("swaps nothing when the server refuses", async () => {
    const m = await load();
    m.setSession(MINE);
    m.updateSettings("gameplay", "strictMana", false);
    m.createPracticeTable.mockRejectedValue(new Error("503"));
    await expect(m.startPractice()).rejects.toThrow("503");
    expect(m.currentSession()?.token).toBe("my-real-seat");
    expect(storedSettings().strictMana).toBe(false);
    expect(localStorage.getItem(m.RECORD_KEY)).toBeNull();
    expect(get(m.practiceTable)).toBeNull();
  });
});

describe("leaving the practice table", () => {
  it("returns a signed-in person to their own session when the one practice replaced has run out (ADR 0110 §1 item 6)", async () => {
    const m = await load();
    const later = (ms: number) => new Date(Date.now() + ms).toISOString();
    const now = new Date().toISOString();
    const identity: Session = {
      token: "my-identity",
      expiresAt: later(30 * 86_400_000),
      principal: {
        role: "identified",
        user_id: "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b",
        issued_at: now,
        expires_at: later(30 * 86_400_000),
      },
    };
    const admin: Session = {
      token: "admin-tok",
      expiresAt: later(3_600_000),
      principal: { role: "admin", issued_at: now, expires_at: later(3_600_000) },
    };
    m.setSession(identity);
    // The admin token over a sign-in keeps the sign-in aside.
    m.setSession(admin);
    m.installPracticeExits();
    m.route.set({ name: "practice" });
    m.createPracticeTable.mockResolvedValue(PRACTICE);
    const id = await m.startPractice();
    m.route.set({ name: "game", gameID: id });

    // The admin session runs out during the tutorial.
    vi.useFakeTimers({ toFake: ["Date"] });
    try {
      vi.setSystemTime(Date.now() + 2 * 3_600_000);
      m.route.set({ name: "lobby" });
    } finally {
      vi.useRealTimers();
    }

    expect(m.currentSession()?.token).toBe("my-identity");
    // The leave call is what puts the cookie back, to the sign-in.
    expect(m.leavePracticeTable).toHaveBeenCalledWith(
      "practice-game",
      "practice-token",
      "my-identity",
      { keepalive: undefined },
    );
  });

  it("Leave (a route change) restores the settings and the session, and abandons the table", async () => {
    const m = await openPractice();
    // A change the player made DURING the tutorial, to a setting the
    // tutorial does not own, is theirs to keep.
    m.updateSettings("display", "theme", "light");

    m.route.set({ name: "lobby" });

    expect(storedSettings()).toEqual(MY_SETTINGS);
    expect(get(m.settings).display.theme).toBe("light");
    expect(m.currentSession()?.token).toBe("my-real-seat");
    expect(JSON.parse(localStorage.getItem(SESSION_KEY) ?? "{}").token).toBe("my-real-seat");
    expect(localStorage.getItem(m.RECORD_KEY)).toBeNull();
    expect(get(m.practiceTable)).toBeNull();
    expect(m.leavePracticeTable).toHaveBeenCalledWith(
      "practice-game",
      "practice-token",
      "my-real-seat",
      { keepalive: undefined },
    );
  });

  it("a closed tab (pagehide) restores before the page goes, and leaves with keepalive", async () => {
    const m = await openPractice();
    window.dispatchEvent(new Event("pagehide"));

    expect(storedSettings()).toEqual(MY_SETTINGS);
    expect(JSON.parse(localStorage.getItem(SESSION_KEY) ?? "{}").token).toBe("my-real-seat");
    expect(m.leavePracticeTable).toHaveBeenCalledWith(
      "practice-game",
      "practice-token",
      "my-real-seat",
      { keepalive: true },
    );
  });

  it("restores a forced setting after an abrupt exit, on the next page load", async () => {
    await openPractice();
    // The tab dies here: no route change, no pagehide. Disk holds the
    // forced settings and the record.
    expect(storedSettings()).toEqual({
      strictMana: true,
      passMode: "manual",
      tableLayout: "quadrant",
      cardSize: "medium",
    });

    vi.resetModules();
    vi.clearAllMocks();
    const next = await load();
    next.route.set({ name: "game", gameID: "practice-game" });
    next.recoverPractice(Date.now() + next.STALE_MS);

    expect(storedSettings()).toEqual(MY_SETTINGS);
    expect(get(next.settings).gameplay.strictMana).toBe(false);
    expect(next.currentSession()?.token).toBe("my-real-seat");
    expect(localStorage.getItem(next.RECORD_KEY)).toBeNull();
    expect(next.leavePracticeTable).toHaveBeenCalledWith(
      "practice-game",
      "practice-token",
      "my-real-seat",
      { keepalive: undefined },
    );
    // A page that was opening the dead table goes to the lobby.
    expect(next.navigate).toHaveBeenCalledWith("#/lobby");
  });

  it("a reload after pagehide lands on the lobby, not the abandoned table", async () => {
    await openPractice();
    window.dispatchEvent(new Event("pagehide"));

    vi.resetModules();
    const next = await load();
    next.route.set({ name: "game", gameID: "practice-game" });
    next.recoverPractice();
    expect(next.navigate).toHaveBeenCalledWith("#/lobby");
  });

  it("a player who signed out during the tutorial stays signed out", async () => {
    const m = await openPractice();
    m.setSession(null);
    m.route.set({ name: "login" });

    expect(storedSettings()).toEqual(MY_SETTINGS);
    expect(m.currentSession()).toBeNull();
    expect(m.leavePracticeTable).toHaveBeenCalledWith("practice-game", "practice-token", "", {
      keepalive: undefined,
    });
  });

  it("an expired previous session is not restored", async () => {
    const m = await load();
    m.setSession(session("old", { expiresAt: new Date(Date.now() + 50).toISOString() }));
    m.installPracticeExits();
    m.route.set({ name: "practice" });
    m.createPracticeTable.mockResolvedValue(PRACTICE);
    await m.startPractice();
    m.route.set({ name: "game", gameID: "practice-game" });

    vi.useFakeTimers({ now: Date.now() + 60_000 });
    try {
      m.route.set({ name: "lobby" });
    } finally {
      vi.useRealTimers();
    }
    expect(m.currentSession()).toBeNull();
    expect(m.leavePracticeTable.mock.calls[0][2]).toBe("");
  });
});

describe("more than one tab", () => {
  it("a page load leaves a freshly stamped practice game to the tab running it", async () => {
    await openPractice();
    vi.resetModules();
    vi.clearAllMocks();
    const other = await load();
    other.route.set({ name: "lobby" });
    other.recoverPractice();

    expect(localStorage.getItem(other.RECORD_KEY)).not.toBeNull();
    expect(other.leavePracticeTable).not.toHaveBeenCalled();
    expect(get(other.practiceTable)).toBeNull();

    // And this tab's own route changes do not end it either.
    other.installPracticeExits();
    other.route.set({ name: "catalog" });
    expect(other.leavePracticeTable).not.toHaveBeenCalled();
  });

  it("a page load opening the practice game adopts it, so its exit restores", async () => {
    await openPractice();
    vi.resetModules();
    vi.clearAllMocks();
    const restored = await load();
    restored.route.set({ name: "game", gameID: "practice-game" });
    restored.recoverPractice();
    restored.installPracticeExits();
    expect(get(restored.practiceTable)).toEqual({ gameID: "practice-game" });

    restored.route.set({ name: "lobby" });
    expect(storedSettings()).toEqual(MY_SETTINGS);
    expect(restored.currentSession()?.token).toBe("my-real-seat");
  });
});

describe("the forced values", () => {
  // ADR 0118 owner decision 7: the tutorial teaches the table the
  // player will meet, so it forces strict payment ON (it forced it off
  // before ADR 0118). A player who turned strict off gets it back off
  // when they leave (the restore tests above use exactly that player).
  it("force strict payment on, and the other three as ADR 0076 §2.2 says", async () => {
    const m = await load();
    expect(m.FORCED_SETTINGS).toEqual({
      strictMana: true,
      passMode: "manual",
      tableLayout: "quadrant",
      cardSize: "medium",
    });
  });
});

describe("helpers", () => {
  it("withSettings touches only the four tutorial-owned fields", async () => {
    const m = await load();
    const before = get(m.settings);
    const after = m.withSettings(before, m.FORCED_SETTINGS);
    expect(after.gameplay.strictMana).toBe(true);
    expect(after.gameplay.passMode).toBe("manual");
    expect(after.display.tableLayout).toBe("quadrant");
    expect(after.display.cardSize).toBe("medium");
    expect({ ...after.gameplay, strictMana: 0, passMode: 0 }).toEqual({
      ...before.gameplay,
      strictMana: 0,
      passMode: 0,
    });
    expect(after.audio).toEqual(before.audio);
  });

  it("parseRecord rejects anything it cannot restore from", async () => {
    const m = await load();
    expect(m.parseRecord(null)).toBeNull();
    expect(m.parseRecord("not json")).toBeNull();
    expect(m.parseRecord(JSON.stringify({ v: 2 }))).toBeNull();
    const good = {
      v: 1,
      gameID: "g",
      practiceToken: "t",
      saved: MY_SETTINGS,
      previous: null,
      aliveAt: 1,
    };
    expect(m.parseRecord(JSON.stringify(good))).toEqual(good);
    // #2336: a player in the focus layout gets it back after the tutorial.
    const focus = { ...good, saved: { ...MY_SETTINGS, tableLayout: "focus" } };
    expect(m.parseRecord(JSON.stringify(focus))).toEqual(focus);
    expect(
      m.parseRecord(
        JSON.stringify({ ...good, saved: { ...MY_SETTINGS, tableLayout: "sideways" } }),
      ),
    ).toBeNull();
    expect(
      m.parseRecord(JSON.stringify({ ...good, saved: { ...MY_SETTINGS, cardSize: "huge" } })),
    ).toBeNull();
    expect(
      m.parseRecord(JSON.stringify({ ...good, saved: { ...MY_SETTINGS, passMode: "fast" } })),
    ).toBeNull();
  });

  // ADR 0143: a record written before passMode existed saved
  // autoPassPriority. It still restores: off is Manual, on is Smart.
  it("parseRecord reads a record from before ADR 0143", async () => {
    const m = await load();
    const { passMode: _drop, ...rest } = MY_SETTINGS;
    void _drop;
    const old = (autoPassPriority: unknown) =>
      JSON.stringify({
        v: 1,
        gameID: "g",
        practiceToken: "t",
        saved: { ...rest, autoPassPriority },
        previous: null,
        aliveAt: 1,
      });
    expect(m.parseRecord(old(true))?.saved.passMode).toBe("smart");
    expect(m.parseRecord(old(false))?.saved.passMode).toBe("manual");
    expect(m.parseRecord(old("yes"))).toBeNull();
  });

  it("isStale is STALE_MS since the last heartbeat", async () => {
    const m = await load();
    const rec = {
      v: 1 as const,
      gameID: "g",
      practiceToken: "t",
      saved: MY_SETTINGS,
      previous: null,
      aliveAt: 1_000,
    };
    expect(m.isStale(rec, 1_000 + m.STALE_MS - 1)).toBe(false);
    expect(m.isStale(rec, 1_000 + m.STALE_MS)).toBe(true);
  });
});
