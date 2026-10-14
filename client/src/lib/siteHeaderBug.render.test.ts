// @vitest-environment jsdom
//
// #3492: Report a bug or idea in the site header, on every page outside
// the table, for someone signed in with Discord when the server files
// reports. The report says which page it came from, never a game.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { writable } from "svelte/store";

import type { Route } from "./router";

const routeStore = vi.hoisted(() => ({
  value: null as null | import("svelte/store").Writable<Route>,
}));
vi.mock("./router", async (importOriginal) => {
  const orig = await importOriginal<typeof import("./router")>();
  routeStore.value = writable<Route>({ name: "decks" } as Route);
  return { ...orig, route: routeStore.value, navigate: vi.fn() };
});

import SiteHeader from "./components/SiteHeader.svelte";
import { resetBugReportConfig } from "./bugReportConfig";
import { L } from "./labels";
import { setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";

function sess(role: Session["principal"]["role"], userID?: string): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: `${role}-tok`,
    expiresAt: expires,
    principal: {
      role,
      user_id: userID,
      name: "Alice",
      issued_at: new Date().toISOString(),
      expires_at: expires,
    },
  } as Session;
}

let enabled = true;
let configCalls = 0;
let reports: unknown[] = [];

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      let payload: unknown = {};
      if (input === "/bugreport/config") {
        configCalls++;
        payload = { enabled, attachments: false };
      } else if (input === "/bugreport" && init.method === "POST") {
        reports.push(JSON.parse(init.body as string));
        payload = { number: 7, url: "https://github.com/x/y/issues/7" };
      }
      return {
        ok: true,
        status: input === "/bugreport" ? 201 : 200,
        statusText: "OK",
        json: async () => payload,
        clone() {
          return this;
        },
      } as unknown as Response;
    }),
  );
}

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

function bugButton(c: HTMLElement): HTMLButtonElement | null {
  return c.querySelector<HTMLButtonElement>(`button[aria-label="${L.reportABug}"]`);
}

beforeEach(() => {
  enabled = true;
  configCalls = 0;
  reports = [];
  resetBugReportConfig();
  stubServer();
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
});

describe("the site header's bug button", () => {
  it("is offered to someone signed in with Discord", async () => {
    setSession(sess("identified", USER));
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    expect(bugButton(container)).not.toBeNull();
  });

  it("is not offered to a guest, or to the admin token", async () => {
    for (const s of [sess("player"), sess("admin"), null]) {
      setSession(s);
      const { container, destroy } = render(SiteHeader as never, {} as never);
      await settle();
      expect(bugButton(container), s?.principal.role ?? "signed out").toBeNull();
      destroy();
    }
  });

  it("is not offered when the server files no reports", async () => {
    enabled = false;
    setSession(sess("identified", USER));
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    expect(bugButton(container)).toBeNull();
  });

  it("asks the server once per tab, however many pages mount a header", async () => {
    setSession(sess("identified", USER));
    render(SiteHeader as never, {} as never);
    await settle();
    render(SiteHeader as never, {} as never);
    await settle();
    expect(configCalls).toBe(1);
  });

  it("files a report that names the page and no game", async () => {
    setSession(sess("identified", USER));
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    click(bugButton(container)!);
    await settle();

    const dialog = document.querySelector<HTMLElement>('[role="dialog"]')!;
    expect(dialog.textContent).toMatch(/attach page context \(page decks\)/);
    const title = dialog.querySelector<HTMLInputElement>("input[type=text], input:not([type])")!;
    title.value = "The deck grid jumps";
    title.dispatchEvent(new Event("input", { bubbles: true }));
    const desc = dialog.querySelector<HTMLTextAreaElement>("textarea")!;
    desc.value = "Scrolling the deck grid jumps back to the top.";
    desc.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    click(dialog.querySelector<HTMLButtonElement>("button.submit")!);
    await settle();

    expect(reports).toHaveLength(1);
    const ctx = (reports[0] as { context?: Record<string, unknown> }).context;
    expect(ctx).toEqual({ page: "decks" });
  });
});
