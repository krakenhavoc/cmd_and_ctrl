// @vitest-environment jsdom
//
// #4103: an idea attaches a copy of the reporter's settings, scrubbed to
// preferences (settingsForReport). The form says so, lets the reporter
// look at it and untick it, and a bug or a question never sends one.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { writable } from "svelte/store";

import type { Route } from "./router";

vi.mock("./router", async (importOriginal) => {
  const orig = await importOriginal<typeof import("./router")>();
  return { ...orig, route: writable<Route>({ name: "decks" } as Route), navigate: vi.fn() };
});

import SiteHeader from "./components/SiteHeader.svelte";
import { _resetForTests as resetDraft } from "./bugReportDraft.svelte";
import { resetBugReportConfig } from "./bugReportConfig";
import { L } from "./labels";
import { _resetForTests as resetModalLayers } from "./modalLayers";
import { resetSettings, updateSettings } from "./settings";
import { setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

function sess(): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: "tok",
    expiresAt: expires,
    principal: {
      role: "identified",
      user_id: "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b",
      name: "Alice",
      issued_at: new Date().toISOString(),
      expires_at: expires,
    },
  } as Session;
}

let reports: Array<Record<string, unknown>> = [];

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      let payload: unknown = {};
      if (input === "/bugreport/config") {
        payload = { enabled: true, attachments: true };
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

const dialog = (): HTMLElement => document.querySelector<HTMLElement>('[role="dialog"]')!;

function settingsBox(): HTMLInputElement | null {
  for (const label of dialog().querySelectorAll<HTMLLabelElement>("label.context")) {
    if (label.textContent?.includes("attach a copy of your settings")) {
      return label.querySelector<HTMLInputElement>('input[type="checkbox"]');
    }
  }
  return null;
}

async function open(kind: string, c: HTMLElement): Promise<void> {
  click(c.querySelector<HTMLButtonElement>(`button[aria-label="${L.reportABug}"]`)!);
  await settle();
  click(dialog().querySelector<HTMLInputElement>(`input[type="radio"][value="${kind}"]`)!);
  const title = dialog().querySelector<HTMLInputElement>('input[type="text"]')!;
  title.value = "Wider stack panel";
  title.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

async function submit(): Promise<void> {
  click(dialog().querySelector<HTMLButtonElement>("button.submit")!);
  await settle();
}

let c: HTMLElement;

beforeEach(async () => {
  reports = [];
  resetDraft();
  resetModalLayers();
  resetBugReportConfig();
  resetSettings();
  updateSettings("display", "cardSize", "large");
  stubServer();
  setSession(sess());
  c = render(SiteHeader as never, {} as never).container;
  await settle();
});
afterEach(() => {
  cleanup();
  resetDraft();
  resetSettings();
  vi.unstubAllGlobals();
  setSession(null);
});

describe("an idea's settings copy", () => {
  it("is sent, as the form shows it", async () => {
    await open("idea", c);
    expect(dialog().textContent).toContain("a copy of your settings");
    const box = settingsBox()!;
    expect(box.checked).toBe(true);

    click(box.parentElement!.querySelector<HTMLButtonElement>("button.peek")!);
    const shown = JSON.parse(dialog().querySelector("pre.logpeek")!.textContent!);
    expect(shown.display.cardSize).toBe("large");
    expect(shown.help).toEqual({ tipsOff: expect.any(Boolean) });

    await submit();
    expect(reports).toHaveLength(1);
    expect(reports[0].settings).toEqual(shown);
  });

  it("is not sent once unticked", async () => {
    await open("idea", c);
    click(settingsBox()!);
    await submit();
    expect(reports).toHaveLength(1);
    expect(reports[0]).not.toHaveProperty("settings");
  });

  it("is never offered or sent with a bug or a question", async () => {
    for (const kind of ["bug", "question"]) {
      resetDraft();
      cleanup();
      c = render(SiteHeader as never, {} as never).container;
      await settle();
      await open(kind, c);
      expect(settingsBox()).toBeNull();
      await submit();
    }
    expect(reports).toHaveLength(2);
    for (const r of reports) expect(r).not.toHaveProperty("settings");
  });
});
