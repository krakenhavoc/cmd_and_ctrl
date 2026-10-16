// @vitest-environment jsdom
//
// #4029: the bug report can be minimized to a pill instead of cancelled.
// The draft (kind, title, details, screenshots, checkbox choices) lives
// in bugReportDraft.svelte.ts, so it survives the form being unmounted.
// Cancel discards it, a successful send clears it, and a minimized report
// holds no modal layer.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { writable } from "svelte/store";

import type { Route } from "./router";

vi.mock("./router", async (importOriginal) => {
  const orig = await importOriginal<typeof import("./router")>();
  return { ...orig, route: writable<Route>({ name: "decks" } as Route), navigate: vi.fn() };
});

import SiteHeader from "./components/SiteHeader.svelte";
import { _resetForTests as resetDraft, currentBugDraft } from "./bugReportDraft.svelte";
import { resetBugReportConfig } from "./bugReportConfig";
import { L } from "./labels";
import { isModalOpen, _resetForTests as resetModalLayers } from "./modalLayers";
import { setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";

function sess(): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: "tok",
    expiresAt: expires,
    principal: {
      role: "identified",
      user_id: USER,
      name: "Alice",
      issued_at: new Date().toISOString(),
      expires_at: expires,
    },
  } as Session;
}

let reports: Array<Record<string, unknown>> = [];
let created: string[] = [];
let revoked: string[] = [];

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      let payload: unknown = {};
      if (input === "/bugreport/config") {
        payload = { enabled: true, attachments: true };
      } else if (input === "/bugreport" && init.method === "POST") {
        // Screenshots go up as multipart; the JSON body only when there are none.
        reports.push(
          init.body instanceof FormData ? { form: true } : JSON.parse(init.body as string),
        );
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

const dialog = (): HTMLElement | null => document.querySelector<HTMLElement>('[role="dialog"]');
const pill = (c: ParentNode = document): HTMLButtonElement | null =>
  c.querySelector<HTMLButtonElement>('button[aria-label="Restore the bug report draft"]');
const bugButton = (c: ParentNode): HTMLButtonElement | null =>
  c.querySelector<HTMLButtonElement>(`button[aria-label="${L.reportABug}"]`);
const minimizeButton = (): HTMLButtonElement | null =>
  document.querySelector<HTMLButtonElement>('button[aria-label="Minimize the report"]');

function type(el: HTMLInputElement | HTMLTextAreaElement, value: string): void {
  el.value = value;
  el.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

function attach(file: File): void {
  const input = dialog()!.querySelector<HTMLInputElement>('input[type="file"]')!;
  Object.defineProperty(input, "files", { value: [file], configurable: true });
  input.dispatchEvent(new Event("change", { bubbles: true }));
  flushSync();
}

// Opens the form and fills every kind of field the draft keeps.
async function startDraft(c: HTMLElement): Promise<void> {
  click(bugButton(c)!);
  await settle();
  const d = dialog()!;
  click(d.querySelector<HTMLInputElement>('input[type="radio"][value="idea"]')!);
  type(d.querySelector<HTMLInputElement>('input[type="text"]')!, "Zoom jumps");
  type(d.querySelector<HTMLTextAreaElement>("textarea")!, "It jumps when I scroll.");
  attach(new File(["png"], "one.png", { type: "image/png" }));
  click(d.querySelector<HTMLInputElement>('.context input[type="checkbox"]')!);
}

let c: HTMLElement;

beforeEach(async () => {
  reports = [];
  created = [];
  revoked = [];
  let n = 0;
  URL.createObjectURL = vi.fn(() => {
    const u = `blob:test/${++n}`;
    created.push(u);
    return u;
  });
  URL.revokeObjectURL = vi.fn((u: string) => {
    revoked.push(u);
  });
  resetDraft();
  resetModalLayers();
  resetBugReportConfig();
  stubServer();
  setSession(sess());
  c = render(SiteHeader as never, {} as never).container;
  await settle();
});
afterEach(() => {
  cleanup();
  resetDraft();
  vi.unstubAllGlobals();
  setSession(null);
});

describe("minimizing the bug report", () => {
  it("hides the form behind a pill and gives the modal layer back", async () => {
    await startDraft(c);
    expect(isModalOpen()).toBe(true);

    click(minimizeButton()!);
    await settle();

    expect(dialog()).toBeNull();
    expect(isModalOpen()).toBe(false);
    const p = pill(c)!;
    expect(p, "the pill should replace the form").toBeTruthy();
    expect(p.tagName).toBe("BUTTON");
    expect(p.textContent).toMatch(/Bug report draft/);
    // Focus lands on the pill rather than falling to <body>.
    expect(document.activeElement).toBe(p);
  });

  it("restores every field, the screenshot and its blob URL from the pill", async () => {
    await startDraft(c);
    click(minimizeButton()!);
    await settle();
    // The screenshot's blob URL is alive while the form is gone.
    expect(created).toHaveLength(1);
    expect(revoked).toEqual([]);

    click(pill(c)!);
    await settle();

    const d = dialog()!;
    expect(pill(c), "the pill gives way to the form").toBeNull();
    expect(d.querySelector<HTMLInputElement>('input[type="radio"][value="idea"]')!.checked).toBe(
      true,
    );
    expect(d.querySelector<HTMLInputElement>('input[type="text"]')!.value).toBe("Zoom jumps");
    expect(d.querySelector<HTMLTextAreaElement>("textarea")!.value).toBe("It jumps when I scroll.");
    expect(d.querySelector<HTMLInputElement>('.context input[type="checkbox"]')!.checked).toBe(
      false,
    );
    const shot = d.querySelector<HTMLImageElement>(".shots img")!;
    expect(shot.getAttribute("src")).toBe(created[0]);
    expect(revoked).toEqual([]);
    // Back at the title, where the pill's focus went.
    expect(document.activeElement).toBe(d.querySelector('input[type="text"]'));
  });

  it("restores the draft when the bug button is used instead of the pill", async () => {
    await startDraft(c);
    click(minimizeButton()!);
    await settle();

    click(bugButton(c)!);
    await settle();

    expect(dialog()!.querySelector<HTMLInputElement>('input[type="text"]')!.value).toBe(
      "Zoom jumps",
    );
  });

  it("keeps the draft when the screen that opened it goes away", async () => {
    await startDraft(c);
    click(minimizeButton()!);
    await settle();
    cleanup();

    // Another page's header, later: same draft, same pill.
    c = render(SiteHeader as never, {} as never).container;
    await settle();
    click(pill(c)!);
    await settle();
    expect(dialog()!.querySelector<HTMLInputElement>('input[type="text"]')!.value).toBe(
      "Zoom jumps",
    );
  });
});

describe("ending the draft", () => {
  it("Cancel discards it and frees the screenshot", async () => {
    await startDraft(c);
    click(dialog()!.querySelector<HTMLButtonElement>("button.cancel")!);
    await settle();

    expect(dialog()).toBeNull();
    expect(pill(c)).toBeNull();
    expect(currentBugDraft()).toBeNull();
    expect(revoked).toEqual(created);

    click(bugButton(c)!);
    await settle();
    expect(dialog()!.querySelector<HTMLInputElement>('input[type="text"]')!.value).toBe("");
    expect(dialog()!.querySelectorAll(".shots img")).toHaveLength(0);
  });

  it("a successful send clears it and frees the screenshot", async () => {
    await startDraft(c);
    click(dialog()!.querySelector<HTMLButtonElement>("button.submit")!);
    await settle();

    expect(reports).toHaveLength(1);
    expect(dialog()!.textContent).toMatch(/Sent/);
    expect(currentBugDraft()).toBeNull();
    expect(revoked).toEqual(created);

    click(dialog()!.querySelector<HTMLButtonElement>("button.submit")!);
    await settle();
    expect(dialog()).toBeNull();
    expect(pill(c)).toBeNull();

    click(bugButton(c)!);
    await settle();
    expect(dialog()!.querySelector<HTMLInputElement>('input[type="text"]')!.value).toBe("");
  });

  it("a failed send keeps the draft, so it can be minimized and tried again", async () => {
    await startDraft(c);
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("offline");
      }),
    );
    click(dialog()!.querySelector<HTMLButtonElement>("button.submit")!);
    await settle();

    expect(currentBugDraft()).not.toBeNull();
    expect(revoked).toEqual([]);
    expect(dialog()!.querySelector('[role="alert"]')!.textContent).toMatch(/offline/);
  });
});

describe("the capture", () => {
  it("is the one taken when the form first opened: a restore reuses the draft", async () => {
    click(bugButton(c)!);
    await settle();
    const first = currentBugDraft()!;
    click(minimizeButton()!);
    await settle();
    click(pill(c)!);
    await settle();
    expect(currentBugDraft(), "a restore reuses the same draft and its capture").toBe(first);
  });
});
