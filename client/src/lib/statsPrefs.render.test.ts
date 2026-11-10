// @vitest-environment jsdom
//
// statsPrefs.render.test.ts — ADR 0217 §2 on the client: the Settings
// "Stats" switch, the opt-in card on #/stats, and the admin's switches
// on Accounts. The routes are a stubbed fetch.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";
import { get } from "svelte/store";

import StatsSettings from "./components/StatsSettings.svelte";
import AdvancedStatsOptIn from "./components/AdvancedStatsOptIn.svelte";
import AccountAdvancedStats from "./components/admin/AccountAdvancedStats.svelte";
import AdvancedStatsDefault from "./components/admin/AdvancedStatsDefault.svelte";
import { L } from "./labels";
import {
  STATS_PREFS_OFF,
  loadStatsPrefs,
  statsPrefs,
  type AdminAccountStats,
  type AdminStatsDefault,
  type StatsPrefs,
} from "./statsPrefs";
import { session, type Session } from "./session";
import { cleanup, click, render } from "./test/render.svelte";

const person = (): Session =>
  ({
    token: "tok",
    expiresAt: "2999-01-01T00:00:00Z",
    principal: { role: "player", user_id: "u1", game_id: "g1", player_id: "p1" },
  }) as unknown as Session;

let server: { prefs: StatsPrefs; calls: { url: string; method: string; body: unknown }[] };

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

beforeEach(() => {
  session.set(person());
  statsPrefs.set(STATS_PREFS_OFF);
  server = {
    prefs: {
      available: true,
      advanced: false,
      choice: "default",
      default: false,
      share_pods: false,
    },
    calls: [],
  };
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init?: RequestInit) => {
      const method = init?.method ?? "GET";
      const body = init?.body ? JSON.parse(String(init.body)) : undefined;
      server.calls.push({ url: String(url), method, body });
      if (String(url).includes("/admin/users/")) {
        const on = (body as { on: boolean }).on;
        return jsonResponse(200, {
          user_id: "u2",
          choice: on ? "on" : "off",
          set_by: "admin",
          advanced: on,
        });
      }
      if (String(url).includes("/admin/stats/advanced-default")) {
        return jsonResponse(200, { on: (body as { on: boolean }).on, stored: true, updated_at: 1 });
      }
      if (method === "PUT") {
        const on = (body as { advanced: boolean }).advanced;
        server.prefs = { ...server.prefs, advanced: on, choice: on ? "on" : "off", set_by: "self" };
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

describe("Settings → Stats", () => {
  it("is off by default and turns advanced stats on", async () => {
    await loadStatsPrefs();
    const r = render(StatsSettings, {});
    flushSync();
    expect(r.container.textContent).toContain(L.advancedStatsSwitch);
    const box = r.container.querySelector<HTMLInputElement>("input[type=checkbox]")!;
    expect(box.checked).toBe(false);
    expect(r.container.textContent).toContain("Off until you turn it on.");
    click(box);
    await settle();
    expect(server.calls.at(-1)).toMatchObject({ method: "PUT", body: { advanced: true } });
    expect(get(statsPrefs).advanced).toBe(true);
    expect(box.checked).toBe(true);
    expect(r.container.textContent).toContain("You turned this on.");
  });

  it("lets a person switch off what an admin turned on", async () => {
    server.prefs = {
      available: true,
      advanced: true,
      choice: "on",
      set_by: "admin",
      default: false,
      share_pods: false,
    };
    await loadStatsPrefs();
    const r = render(StatsSettings, {});
    flushSync();
    expect(r.container.textContent).toContain("An admin turned this on for you");
    click(r.container.querySelector("input[type=checkbox]")!);
    await settle();
    expect(server.calls.at(-1)).toMatchObject({ method: "PUT", body: { advanced: false } });
    expect(get(statsPrefs)).toMatchObject({ advanced: false, set_by: "self" });
  });

  it("never asks the server for a guest", async () => {
    session.set({
      token: "g",
      expiresAt: "2999-01-01T00:00:00Z",
      principal: { role: "player", game_id: "g1", player_id: "p1" },
    } as unknown as Session);
    expect(await loadStatsPrefs()).toEqual(STATS_PREFS_OFF);
    expect(server.calls).toHaveLength(0);
  });
});

describe("the opt-in card on #/stats", () => {
  it("offers advanced stats while they are off, then says they are on", async () => {
    await loadStatsPrefs();
    const r = render(AdvancedStatsOptIn, {});
    flushSync();
    const button = [...r.container.querySelectorAll("button")].find(
      (b) => b.textContent?.trim() === L.advancedStatsOptIn,
    );
    expect(button).toBeTruthy();
    click(button!);
    await settle();
    expect(server.calls.at(-1)).toMatchObject({ method: "PUT", body: { advanced: true } });
    expect(r.container.querySelector("[data-testid=advanced-stats-opt-in]")).toBeNull();
    expect(r.container.querySelector("[data-testid=advanced-stats-on]")).not.toBeNull();
  });

  it("draws nothing on a server that keeps no setting", () => {
    const r = render(AdvancedStatsOptIn, {});
    flushSync();
    expect(r.container.textContent?.trim()).toBe("");
  });
});

describe("the admin's switches", () => {
  it("switches one person on, named in its label", async () => {
    const changed: AdminAccountStats[] = [];
    const r = render(AccountAdvancedStats, {
      name: "Bea",
      row: { user_id: "u2", choice: "default", advanced: false } as AdminAccountStats,
      onchange: (row: AdminAccountStats) => changed.push(row),
      onforbidden: () => {},
    });
    flushSync();
    expect(r.container.textContent).toContain("Default (off)");
    const box = r.container.querySelector<HTMLInputElement>(
      `input[aria-label="${L.accountAdvancedStats("Bea")}"]`,
    )!;
    click(box);
    await settle();
    expect(server.calls.at(-1)).toMatchObject({
      url: "/admin/users/u2/advanced-stats",
      method: "PUT",
      body: { on: true },
    });
    expect(changed[0]).toMatchObject({ choice: "on", set_by: "admin" });
  });

  it("stores the default", async () => {
    const changed: AdminStatsDefault[] = [];
    const r = render(AdvancedStatsDefault, {
      value: { on: false, stored: false },
      onchange: (d: AdminStatsDefault) => changed.push(d),
      onforbidden: () => {},
    });
    flushSync();
    expect(r.container.textContent).toContain(L.advancedStatsDefault);
    expect(r.container.textContent).toContain("never set");
    click(r.container.querySelector("input[type=checkbox]")!);
    await settle();
    expect(server.calls.at(-1)).toMatchObject({
      url: "/admin/stats/advanced-default",
      method: "PUT",
      body: { on: true },
    });
    expect(changed[0]).toEqual({ on: true, stored: true, updated_at: 1 });
  });
});
