// @vitest-environment jsdom
//
// notifications.render.test.ts — ADR 0162 §1 and §4 on the client: the
// Settings "Notifications" tab (three switches, the Discord-refused
// hint) and the one-time nudge at the table, whose answer lives in the
// synced settings. GET and PUT /me/notifications are a stubbed fetch.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";
import { get } from "svelte/store";

import NotificationSettings from "./components/NotificationSettings.svelte";
import NotificationsNudge from "./components/NotificationsNudge.svelte";
import {
  DM_FAILED_HINT,
  NOTIFICATIONS_OFF,
  NUDGE_ID,
  loadNotifications,
  notificationPrefs,
  shouldNudge,
  type NotificationPrefs,
} from "./notifications";
import { resetSettings, settings } from "./settings";
import { session, type Session } from "./session";
import { cleanup, click, render } from "./test/render.svelte";

const person = (): Session =>
  ({
    token: "tok",
    expiresAt: "2999-01-01T00:00:00Z",
    principal: { role: "player", user_id: "u1", game_id: "g1", player_id: "p1" },
  }) as unknown as Session;

const ON: NotificationPrefs = { ...NOTIFICATIONS_OFF, available: true };

// server is the stubbed /me/notifications: its current switches, and
// every PUT body it was sent.
let server: { status: number; prefs: NotificationPrefs; puts: Record<string, boolean>[] };

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

beforeEach(() => {
  resetSettings();
  session.set(person());
  notificationPrefs.set(NOTIFICATIONS_OFF);
  server = { status: 200, prefs: { ...ON }, puts: [] };
  vi.stubGlobal(
    "fetch",
    vi.fn(async (_url: string, init?: RequestInit) => {
      if (server.status !== 200) {
        return jsonResponse(server.status, { error: "not available", available: false });
      }
      if (init?.method === "PUT") {
        const body = JSON.parse(String(init.body)) as Record<string, boolean>;
        server.puts.push(body);
        server.prefs = { ...server.prefs, ...body };
        if (Object.values(body).some(Boolean)) server.prefs.dm_failed = false;
      }
      return jsonResponse(200, server.prefs);
    }),
  );
});

afterEach(() => {
  cleanup();
  session.set(null);
  vi.unstubAllGlobals();
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) await new Promise((r) => setTimeout(r, 0));
  flushSync();
}

describe("shouldNudge", () => {
  it("only when available, every switch off, and not answered before", () => {
    expect(shouldNudge(ON, {})).toBe(true);
    expect(shouldNudge(NOTIFICATIONS_OFF, {})).toBe(false);
    expect(shouldNudge({ ...ON, table_ready: true }, {})).toBe(false);
    expect(shouldNudge(ON, { [NUDGE_ID]: 1 })).toBe(false);
  });
});

describe("loadNotifications", () => {
  it("reads nothing for a session with no user, and stays off", async () => {
    session.set(null);
    expect(await loadNotifications()).toEqual(NOTIFICATIONS_OFF);
    expect(fetch).not.toHaveBeenCalled();
  });

  it("treats a 503 (no bot token) as not available", async () => {
    server.status = 503;
    expect((await loadNotifications()).available).toBe(false);
  });
});

describe("the Notifications tab", () => {
  it("shows the three switches off by default and saves one", async () => {
    await loadNotifications();
    const r = render(NotificationSettings, {});
    const boxes = [...r.container.querySelectorAll<HTMLInputElement>("input[type=checkbox]")];
    expect(boxes).toHaveLength(3);
    expect(boxes.every((b) => !b.checked)).toBe(true);
    expect(r.container.textContent).not.toContain(DM_FAILED_HINT);

    click(boxes[0]);
    await settle();
    expect(server.puts).toEqual([{ turn: true }]);
    expect(get(notificationPrefs).turn).toBe(true);
  });

  it("shows the Discord-refused hint until a switch is turned on", async () => {
    server.prefs = { ...ON, turn: true, dm_failed: true };
    await loadNotifications();
    const r = render(NotificationSettings, {});
    expect(r.container.textContent).toContain(DM_FAILED_HINT);

    const ready = r.container.querySelectorAll<HTMLInputElement>("input[type=checkbox]")[1];
    click(ready);
    await settle();
    expect(r.container.textContent).not.toContain(DM_FAILED_HINT);
  });
});

describe("the nudge at the table", () => {
  it("offers Turn on once, and remembers the answer in the synced settings", async () => {
    const r = render(NotificationsNudge, {});
    await settle();
    expect(r.container.textContent).toContain("Get a Discord DM when it's your turn?");

    const turnOn = [...r.container.querySelectorAll("button")].find(
      (b) => b.textContent === "Turn on",
    );
    click(turnOn!);
    await settle();
    expect(server.puts).toEqual([{ turn: true }]);
    expect(get(settings).help.seen[NUDGE_ID]).toBe(1);
    expect(r.container.textContent).not.toContain("your turn?");
  });

  it("Not now hides it for good without changing a switch", async () => {
    const r = render(NotificationsNudge, {});
    await settle();
    const notNow = [...r.container.querySelectorAll("button")].find(
      (b) => b.textContent === "Not now",
    );
    click(notNow!);
    expect(r.container.textContent).toBe("");
    expect(server.puts).toEqual([]);

    cleanup();
    const again = render(NotificationsNudge, {});
    await settle();
    expect(again.container.textContent).toBe("");
  });

  it("draws nothing when the server has no bot token", async () => {
    server.status = 503;
    const r = render(NotificationsNudge, {});
    await settle();
    expect(r.container.textContent).toBe("");
  });

  it("draws nothing when a switch is already on", async () => {
    server.prefs = { ...ON, table_ready: true };
    const r = render(NotificationsNudge, {});
    await settle();
    expect(r.container.textContent).toBe("");
  });
});
