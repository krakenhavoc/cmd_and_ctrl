// @vitest-environment jsdom
//
// settingsSync.test.ts — a signed-in person's settings on their account
// (ADR 0110 §4, Delivery PR 5). The network is two spies; the session
// and settings stores are the real ones, loaded fresh for each test (a
// new page, as far as module state is concerned).

import { afterEach, beforeEach, describe, expect, it, vi, type Mock } from "vitest";
import { get, writable } from "svelte/store";
import type { Session } from "./session";
import type { Route } from "./router";
import type { PracticeRecord } from "./practiceTable";

const SETTINGS_KEY = "cmdctrl.settings.v1";
const SESSION_KEY = "cmdctrl.session";
const RECORD_KEY = "cmdctrl.practice.v1";
const USER = "11111111-2222-3333-4444-555555555555";

function sessionFor(role: Session["principal"]["role"], userID?: string, token = "tok"): Session {
  const expires = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token,
    expiresAt: expires,
    principal: { role, user_id: userID, issued_at: new Date().toISOString(), expires_at: expires },
  };
}

const SIGNED_IN = sessionFor("identified", USER, "identity-token");

// storedBrowserSettings writes a settings blob this browser already
// holds before the page loads.
function storeBrowserSettings(patch: Record<string, Record<string, unknown>>) {
  const blob: Record<string, unknown> = { __version: 15 };
  for (const [g, v] of Object.entries(patch)) blob[g] = v;
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(blob));
}

let uninstall: (() => void) | null = null;

async function load(opts: { session?: Session | null; install?: boolean } = {}) {
  vi.resetModules();
  vi.doMock("./router", () => ({
    route: writable<Route>({ name: "lobby" }),
    navigate: vi.fn(),
  }));
  vi.doMock("./api", () => ({
    fetchAccountSettings: vi.fn(async () => ({ revision: 0 })),
    putAccountSettings: vi.fn(async (revision: number, version: number, body: unknown) => ({
      revision: revision + 1,
      version,
      settings: body,
    })),
    createPracticeTable: vi.fn(),
    leavePracticeTable: vi.fn(async () => {}),
  }));
  if (opts.session !== undefined && opts.session !== null) {
    localStorage.setItem(SESSION_KEY, JSON.stringify(opts.session));
  }
  const sync = await import("./settingsSync");
  const settingsMod = await import("./settings");
  const sessionMod = await import("./session");
  const practice = await import("./practiceTable");
  const api = await import("./api");
  const m = {
    ...sync,
    settings: settingsMod.settings,
    updateSettings: settingsMod.updateSettings,
    SETTINGS_VERSION: settingsMod.SETTINGS_VERSION,
    setSession: sessionMod.setSession,
    LobbyApiError: sessionMod.LobbyApiError,
    practice,
    fetch: api.fetchAccountSettings as unknown as Mock,
    put: api.putAccountSettings as unknown as Mock,
  };
  if (opts.install !== false) uninstall = m.installSettingsSync();
  return m;
}

// settle lets every pending promise and zero-delay timer run.
async function settle() {
  await vi.advanceTimersByTimeAsync(0);
}

// lastPut returns the body of the newest PUT: its revision, version
// and settings.
function lastPut(put: Mock) {
  const call = put.mock.calls.at(-1);
  if (!call) throw new Error("no PUT was sent");
  return {
    revision: call[0] as number,
    version: call[1] as number,
    body: call[2] as Record<string, Record<string, unknown>>,
  };
}

beforeEach(() => {
  vi.useFakeTimers();
  localStorage.clear();
  sessionStorage.clear();
});

afterEach(() => {
  uninstall?.();
  uninstall = null;
  // End a practice table an earlier module instance is still running,
  // before the next test wipes the storage it would write back into.
  window.dispatchEvent(new Event("pagehide"));
  vi.useRealTimers();
  vi.clearAllMocks();
});

describe("guests and other sessions without a person", () => {
  it("a guest seat never talks to the account", async () => {
    const m = await load({ session: sessionFor("player") });
    await settle();
    m.updateSettings("display", "theme", "light");
    await vi.advanceTimersByTimeAsync(10_000);
    expect(m.fetch).not.toHaveBeenCalled();
    expect(m.put).not.toHaveBeenCalled();
  });

  it("the admin token, or no session at all, never talks to the account", async () => {
    const m = await load({ session: sessionFor("admin") });
    await settle();
    m.setSession(null);
    m.updateSettings("display", "theme", "light");
    await vi.advanceTimersByTimeAsync(10_000);
    expect(m.fetch).not.toHaveBeenCalled();
    expect(m.put).not.toHaveBeenCalled();
  });

  it("a 403 from the server leaves the settings browser-only", async () => {
    const m = await load({ install: false });
    m.fetch.mockRejectedValue(new m.LobbyApiError(403, "not a person"));
    m.setSession(SIGNED_IN);
    uninstall = m.installSettingsSync();
    await settle();
    m.updateSettings("display", "theme", "light");
    await vi.advanceTimersByTimeAsync(10_000);
    expect(m.fetch).toHaveBeenCalledOnce();
    expect(m.put).not.toHaveBeenCalled();
  });
});

describe("signing in", () => {
  it("with no account copy, uploads this browser's per-person settings as the first copy", async () => {
    storeBrowserSettings({
      display: { theme: "light", cardSize: "large" },
      audio: { masterVolume: 12 },
    });
    const m = await load({ session: SIGNED_IN });
    await settle();
    expect(m.fetch).toHaveBeenCalledOnce();
    expect(m.put).toHaveBeenCalledOnce();
    const sent = lastPut(m.put);
    expect(sent.revision).toBe(0);
    expect(sent.version).toBe(m.SETTINGS_VERSION);
    expect(sent.body.display.theme).toBe("light");
    // Per-device fields never leave the browser.
    expect(sent.body.display).not.toHaveProperty("cardSize");
    expect(sent.body).not.toHaveProperty("audio");
    expect(get(m.settingsSyncToast)).toBeNull();
  });

  it("downloads on sign-in, not only on page load", async () => {
    const m = await load({ session: null });
    await settle();
    expect(m.fetch).not.toHaveBeenCalled();
    m.setSession(SIGNED_IN);
    await settle();
    expect(m.fetch).toHaveBeenCalledOnce();
  });

  it("a token swap for the same person is not a sign-in", async () => {
    const m = await load({ session: SIGNED_IN });
    await settle();
    m.setSession(sessionFor("player", USER, "seat-token"));
    await settle();
    expect(m.fetch).toHaveBeenCalledOnce();
  });

  it("the account's copy wins over different browser values, and the toast offers the browser's back", async () => {
    storeBrowserSettings({
      display: { theme: "light", cardSize: "large" },
      gameplay: { strictMana: true },
    });
    const m = await load({ install: false });
    m.fetch.mockResolvedValue({
      version: m.SETTINGS_VERSION,
      revision: 4,
      settings: { display: { theme: "high-contrast" }, gameplay: { strictMana: false } },
    });
    m.setSession(SIGNED_IN);
    uninstall = m.installSettingsSync();
    await settle();

    const s = get(m.settings);
    expect(s.display.theme).toBe("high-contrast");
    expect(s.gameplay.strictMana).toBe(false);
    // The device's own fields are untouched.
    expect(s.display.cardSize).toBe("large");
    expect(get(m.settingsSyncToast)).not.toBeNull();
    // Applying the account's copy is not a change to upload.
    await vi.advanceTimersByTimeAsync(5_000);
    expect(m.put).not.toHaveBeenCalled();

    // "Keep this browser's instead" puts them back and uploads them.
    m.keepBrowserSettings();
    await settle();
    expect(get(m.settingsSyncToast)).toBeNull();
    expect(get(m.settings).display.theme).toBe("light");
    expect(get(m.settings).gameplay.strictMana).toBe(true);
    const sent = lastPut(m.put);
    expect(sent.revision).toBe(4);
    expect(sent.body.display.theme).toBe("light");
    expect(sent.body.gameplay.strictMana).toBe(true);
  });

  it("dismissing the toast keeps the account's settings", async () => {
    storeBrowserSettings({ display: { theme: "light" } });
    const m = await load({ install: false });
    m.fetch.mockResolvedValue({
      version: m.SETTINGS_VERSION,
      revision: 2,
      settings: { display: { theme: "dark" } },
    });
    m.setSession(SIGNED_IN);
    uninstall = m.installSettingsSync();
    await settle();
    m.dismissSettingsSyncToast();
    expect(get(m.settingsSyncToast)).toBeNull();
    expect(get(m.settings).display.theme).toBe("dark");
    await vi.advanceTimersByTimeAsync(5_000);
    expect(m.put).not.toHaveBeenCalled();
  });

  it("an account copy that matches this browser shows no toast and uploads nothing", async () => {
    const m = await load({ install: false });
    const { syncedSubset } = await import("./settings");
    m.fetch.mockResolvedValue({
      version: m.SETTINGS_VERSION,
      revision: 7,
      settings: syncedSubset(get(m.settings)),
    });
    m.setSession(SIGNED_IN);
    uninstall = m.installSettingsSync();
    await settle();
    expect(get(m.settingsSyncToast)).toBeNull();
    await vi.advanceTimersByTimeAsync(5_000);
    expect(m.put).not.toHaveBeenCalled();
  });

  it("signing out keeps the values this browser has, and hides the toast", async () => {
    storeBrowserSettings({ display: { theme: "light" } });
    const m = await load({ install: false });
    m.fetch.mockResolvedValue({
      version: m.SETTINGS_VERSION,
      revision: 2,
      settings: { display: { theme: "dark" } },
    });
    m.setSession(SIGNED_IN);
    uninstall = m.installSettingsSync();
    await settle();
    m.setSession(null);
    expect(get(m.settingsSyncToast)).toBeNull();
    expect(get(m.settings).display.theme).toBe("dark");
  });

  it("a copy from a newer client is applied, and this tab stops writing", async () => {
    const m = await load({ install: false });
    m.fetch.mockResolvedValue({
      version: m.SETTINGS_VERSION + 1,
      revision: 3,
      settings: { display: { theme: "light" }, futureGroup: { x: 1 } },
    });
    m.setSession(SIGNED_IN);
    uninstall = m.installSettingsSync();
    await settle();
    expect(get(m.settings).display.theme).toBe("light");
    expect(get(m.settingsSyncToast)).toBeNull();
    m.updateSettings("display", "theme", "dark");
    await vi.advanceTimersByTimeAsync(5_000);
    expect(m.put).not.toHaveBeenCalled();
  });
});

describe("uploading a change", () => {
  async function signedIn() {
    const m = await load({ install: false });
    m.fetch.mockResolvedValue({ version: m.SETTINGS_VERSION, revision: 1, settings: {} });
    m.setSession(SIGNED_IN);
    uninstall = m.installSettingsSync();
    await settle();
    m.put.mockClear();
    return m;
  }

  it("is debounced: several quick changes are one PUT, a second later", async () => {
    const m = await signedIn();
    m.updateSettings("display", "theme", "light");
    await vi.advanceTimersByTimeAsync(400);
    m.updateSettings("display", "hoverDelayMs", 600);
    await vi.advanceTimersByTimeAsync(900);
    expect(m.put).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(200);
    expect(m.put).toHaveBeenCalledOnce();
    const sent = lastPut(m.put);
    expect(sent.revision).toBe(1);
    expect(sent.body.display.theme).toBe("light");
    expect(sent.body.display.hoverDelayMs).toBe(600);

    // The next change rides the new revision.
    m.updateSettings("display", "theme", "dark");
    await vi.advanceTimersByTimeAsync(1_000);
    expect(lastPut(m.put).revision).toBe(2);
  });

  it("a per-device change uploads nothing", async () => {
    const m = await signedIn();
    m.updateSettings("display", "cardSize", "small");
    m.updateSettings("audio", "masterVolume", 3);
    m.updateSettings("accessibility", "textScale", 1.5);
    await vi.advanceTimersByTimeAsync(5_000);
    expect(m.put).not.toHaveBeenCalled();
  });

  it("a failed upload retries later and never throws into the UI", async () => {
    const m = await signedIn();
    m.put.mockRejectedValueOnce(new TypeError("Failed to fetch"));
    m.updateSettings("display", "theme", "light");
    await vi.advanceTimersByTimeAsync(1_000);
    expect(m.put).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(m.RETRY_MIN_MS);
    expect(m.put).toHaveBeenCalledTimes(2);
    expect(lastPut(m.put).body.display.theme).toBe("light");
  });

  it("a failed upload retries at once when the browser comes back online", async () => {
    const m = await signedIn();
    m.put.mockRejectedValueOnce(new TypeError("offline"));
    m.updateSettings("display", "theme", "light");
    await vi.advanceTimersByTimeAsync(1_000);
    window.dispatchEvent(new Event("online"));
    await settle();
    expect(m.put).toHaveBeenCalledTimes(2);
  });

  it("a pending change is flushed on pagehide, with keepalive", async () => {
    const m = await signedIn();
    m.updateSettings("display", "theme", "light");
    window.dispatchEvent(new Event("pagehide"));
    await settle();
    expect(m.put).toHaveBeenCalledOnce();
    expect(m.put.mock.calls[0][3]).toEqual({ keepalive: true });
  });

  it("a 412 merges: this browser's own change survives, the account wins the rest", async () => {
    const m = await signedIn();
    // Another device changed theme AND hoverDelayMs; this browser
    // changed stackStyle and hoverDelayMs.
    m.put.mockRejectedValueOnce(
      new m.LobbyApiError(412, "changed somewhere else", undefined, undefined, {
        version: m.SETTINGS_VERSION,
        revision: 5,
        settings: { display: { theme: "light", hoverDelayMs: 900 } },
      }),
    );
    m.updateSettings("display", "stackStyle", "fan");
    m.updateSettings("display", "hoverDelayMs", 100);
    await vi.advanceTimersByTimeAsync(1_000);
    expect(m.put).toHaveBeenCalledTimes(1);

    const s = get(m.settings);
    expect(s.display.theme).toBe("light"); // only the account changed it
    expect(s.display.stackStyle).toBe("fan"); // only this browser changed it
    expect(s.display.hoverDelayMs).toBe(900); // both did: the account wins

    // The merged copy goes up over the new revision.
    await vi.advanceTimersByTimeAsync(1_000);
    expect(m.put).toHaveBeenCalledTimes(2);
    const sent = lastPut(m.put);
    expect(sent.revision).toBe(5);
    expect(sent.body.display).toMatchObject({
      theme: "light",
      stackStyle: "fan",
      hoverDelayMs: 900,
    });
  });

  it("a 409 (a newer client saved the copy) takes the account's copy and stops writing", async () => {
    const m = await signedIn();
    m.put.mockRejectedValueOnce(
      new m.LobbyApiError(409, "reload", undefined, undefined, {
        version: m.SETTINGS_VERSION + 1,
        revision: 9,
        settings: { display: { theme: "high-contrast" } },
      }),
    );
    m.updateSettings("display", "theme", "light");
    await vi.advanceTimersByTimeAsync(1_000);
    expect(get(m.settings).display.theme).toBe("high-contrast");
    m.updateSettings("display", "theme", "dark");
    await vi.advanceTimersByTimeAsync(10_000);
    expect(m.put).toHaveBeenCalledTimes(1);
  });
});

describe("the practice table's forced settings never reach the account", () => {
  function practiceRecord(saved: PracticeRecord["saved"], aliveAt = Date.now()): PracticeRecord {
    return {
      v: 1,
      gameID: "practice-game",
      practiceToken: "practice-token",
      saved,
      previous: SIGNED_IN,
      aliveAt,
    };
  }
  const SAVED = {
    strictMana: true,
    autoPassPriority: true,
    tableLayout: "row",
    cardSize: "large",
  } as const;

  it("while a practice table is active, a change uploads the saved values for the forced keys", async () => {
    const m = await load({ install: false });
    m.fetch.mockResolvedValue({ version: m.SETTINGS_VERSION, revision: 1, settings: {} });
    m.setSession(SIGNED_IN);
    uninstall = m.installSettingsSync();
    await settle();
    m.put.mockClear();

    m.updateSettings("gameplay", "strictMana", true);
    m.updateSettings("gameplay", "autoPassPriority", true);
    await vi.advanceTimersByTimeAsync(1_000);
    m.put.mockClear();

    // Open the tutorial: it forces strictMana and autoPassPriority off,
    // and the practice seat's session keeps the same person.
    m.practice.installPracticeExits();
    vi.mocked((await import("./api")).createPracticeTable).mockResolvedValue({
      ...sessionFor("player", USER, "practice-token"),
      gameID: "practice-game",
      playerID: "seat-0",
    });
    await m.practice.startPractice();
    expect(get(m.settings).gameplay.strictMana).toBe(false);
    await vi.advanceTimersByTimeAsync(5_000);
    // Forcing the tutorial's values is not a change to the account.
    expect(m.put).not.toHaveBeenCalled();

    // A change during the tutorial goes up with the player's own
    // values for the forced keys.
    m.updateSettings("display", "theme", "light");
    await vi.advanceTimersByTimeAsync(1_000);
    expect(m.put).toHaveBeenCalledOnce();
    const sent = lastPut(m.put);
    expect(sent.body.display.theme).toBe("light");
    expect(sent.body.gameplay.strictMana).toBe(true);
    expect(sent.body.gameplay.autoPassPriority).toBe(true);
    for (const call of m.put.mock.calls) {
      const body = call[2] as Record<string, Record<string, unknown>>;
      expect(body.gameplay.strictMana).toBe(true);
    }
  });

  it("a tutorial that crashed mid-way cannot put its forced values on the account", async () => {
    // Disk after the crash: the forced values, and the record of the
    // player's own. No code has run to restore them.
    storeBrowserSettings({
      gameplay: { strictMana: false, autoPassPriority: false },
      display: { tableLayout: "quadrant", cardSize: "medium" },
    });
    localStorage.setItem(RECORD_KEY, JSON.stringify(practiceRecord(SAVED)));
    const m = await load({ session: SIGNED_IN });
    await settle();
    expect(m.put).toHaveBeenCalledOnce();
    const sent = lastPut(m.put);
    expect(sent.body.gameplay.strictMana).toBe(true);
    expect(sent.body.gameplay.autoPassPriority).toBe(true);
    expect(sent.body.display).not.toHaveProperty("tableLayout");
    expect(sent.body.display).not.toHaveProperty("cardSize");
  });

  it("a copy that arrives during the tutorial lands in the record, and the forced values stay live", async () => {
    storeBrowserSettings({
      gameplay: { strictMana: false, autoPassPriority: false },
      display: { tableLayout: "quadrant", cardSize: "medium" },
    });
    localStorage.setItem(RECORD_KEY, JSON.stringify(practiceRecord(SAVED)));
    const m = await load({ install: false });
    m.fetch.mockResolvedValue({
      version: m.SETTINGS_VERSION,
      revision: 3,
      settings: {
        gameplay: { strictMana: false, autoPassPriority: true },
        display: { theme: "light" },
      },
    });
    m.setSession(SIGNED_IN);
    uninstall = m.installSettingsSync();
    await settle();

    const live = get(m.settings);
    expect(live.display.theme).toBe("light");
    // Still the tutorial's while it runs.
    expect(live.gameplay.strictMana).toBe(false);
    expect(live.gameplay.autoPassPriority).toBe(false);
    // The restore will put the account's values back.
    expect(m.practice.practiceSaved()).toEqual({
      strictMana: false,
      autoPassPriority: true,
      tableLayout: "row",
      cardSize: "large",
    });
    // The browser's own values differed (strictMana true), so the
    // toast is offered; nothing is uploaded until the player chooses.
    expect(get(m.settingsSyncToast)).not.toBeNull();
    await vi.advanceTimersByTimeAsync(5_000);
    expect(m.put).not.toHaveBeenCalled();
  });
});
