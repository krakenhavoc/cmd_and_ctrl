// @vitest-environment jsdom
//
// myDecks.render.test.ts — ADR 0110 section 6: the #/decks page lists
// saved decks with "N of M cards play as printed", renames and deletes
// (with a confirm), shows the source link, and the lobby's picker
// shows the same count beside each saved deck.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import MyDecks from "../routes/MyDecks.svelte";
import YourDecksPicker from "./components/YourDecksPicker.svelte";
import { sessionFromOAuth, setSession } from "./session";
import { click, cleanup, flushSync, render } from "./test/render.svelte";
import type { MyDeckInfo } from "./myDecks";

const COUNTS = { manual: 2, unreviewed: 0, caveats: 1, automated: 5, no_effect: 2 };

function deck(over: Partial<MyDeckInfo>): MyDeckInfo {
  return {
    id: "d1",
    name: "Atraxa",
    commanders: ["Atraxa, Praetors' Voice"],
    card_count: 100,
    updated_at: "2026-09-19T08:00:00Z",
    coverage: { counts: COUNTS, unknown: 0, as_printed: 7, resolved: 10 },
    ...over,
  };
}

interface Call {
  url: string;
  method: string;
  body?: string;
}

let calls: Call[];
let library: MyDeckInfo[];
let renameStatus = 200;

function respond(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: "",
    json: async () => body,
    clone: () => ({ json: async () => body }),
  } as unknown as Response;
}

beforeEach(() => {
  calls = [];
  renameStatus = 200;
  library = [
    deck({}),
    deck({
      id: "d2",
      name: "Linked",
      source_url: "https://moxfield.com/decks/abc",
      coverage: { counts: COUNTS, unknown: 0, as_printed: 3, resolved: 4 },
    }),
  ];
  setSession(sessionFromOAuth({ token: "tok", expiresAt: "2099-01-01T00:00:00Z", userID: "u1" }));
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init?: RequestInit) => {
      const method = init?.method ?? "GET";
      calls.push({ url, method, body: init?.body as string | undefined });
      if (url === "/me/decks") return respond(200, { decks: library });
      if (url.endsWith("/coverage")) {
        return respond(200, {
          deck_name: "Atraxa",
          source: "text",
          commanders: [],
          counts: COUNTS,
          cards: [
            { name: "Doubling Season", oracle_id: "o1", count: 1, bucket: "manual" },
            { name: "Sol Ring", oracle_id: "o2", count: 1, bucket: "automated" },
          ],
          unknown: [],
          violations: [],
        });
      }
      if (method === "PATCH") {
        if (renameStatus === 409)
          return respond(409, { error: "you already have a deck with that name" });
        return respond(200, { ...library[0], name: JSON.parse(init?.body as string).name });
      }
      if (method === "DELETE") return respond(204, null);
      return respond(404, { error: "nope" });
    }),
  );
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

const button = (c: HTMLElement, text: string): HTMLButtonElement =>
  [...c.querySelectorAll<HTMLButtonElement>("button")].find((b) => b.textContent?.trim() === text)!;

describe("#/decks", () => {
  it("lists each saved deck with its playable count and source link", async () => {
    const { container } = render(MyDecks as never, {} as never);
    await settle();
    const text = container.textContent ?? "";
    expect(text).toContain("7 of 10 cards play as printed");
    expect(text).toContain("3 of 4 cards play as printed");
    expect(text).toContain("1 simplified");
    expect(text).toContain("2 you resolve by hand");
    const link = container.querySelector<HTMLAnchorElement>(
      'a[href="https://moxfield.com/decks/abc"]',
    );
    expect(link?.textContent).toBe("moxfield.com");
  });

  it("opens the full report in place", async () => {
    const { container } = render(MyDecks as never, {} as never);
    await settle();
    click(button(container, "Coverage report"));
    await settle();
    expect(calls.some((c) => c.url === "/me/decks/d1/coverage")).toBe(true);
    expect(container.textContent).toContain("Doubling Season");
  });

  it("renames, and shows the server's message on a conflict", async () => {
    const { container } = render(MyDecks as never, {} as never);
    await settle();
    click(button(container, "Rename"));
    flushSync();
    const input = container.querySelector<HTMLInputElement>("form.rename input")!;
    input.value = "Atraxa Prime";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    click(button(container, "Save"));
    await settle();
    const patch = calls.find((c) => c.method === "PATCH")!;
    expect(patch.url).toBe("/me/decks/d1");
    expect(JSON.parse(patch.body!)).toEqual({ name: "Atraxa Prime" });
    expect(container.textContent).toContain("Atraxa Prime");

    renameStatus = 409;
    click(button(container, "Rename"));
    flushSync();
    const again = container.querySelector<HTMLInputElement>("form.rename input")!;
    again.value = "Linked";
    again.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    click(button(container, "Save"));
    await settle();
    expect(container.textContent).toContain("you already have a deck with that name");
  });

  it("deletes only after a confirm", async () => {
    const { container } = render(MyDecks as never, {} as never);
    await settle();
    click(container.querySelector<HTMLButtonElement>('button[aria-label="delete Atraxa"]')!);
    flushSync();
    expect(calls.some((c) => c.method === "DELETE")).toBe(false);
    expect(container.textContent).toContain("Delete Atraxa?");
    click(button(container, "Delete"));
    await settle();
    expect(calls.find((c) => c.method === "DELETE")?.url).toBe("/me/decks/d1");
    expect(container.textContent).not.toContain("7 of 10 cards play as printed");
    expect(container.textContent).toContain("3 of 4 cards play as printed");
  });

  it("asks a signed-out visitor to sign in without calling the server", async () => {
    setSession(null);
    const { container } = render(MyDecks as never, {} as never);
    await settle();
    expect(container.textContent).toContain("Sign in with Discord");
    expect(calls).toHaveLength(0);
  });
});

describe("the lobby's deck picker", () => {
  it("shows the playable count beside each saved deck", async () => {
    const { container } = render(
      YourDecksPicker as never,
      { gameID: "g1", playerID: "p1" } as never,
    );
    await settle();
    const text = container.textContent ?? "";
    expect(text).toContain("7 of 10 cards play as printed");
    expect(text).toContain("3 of 4 cards play as printed");
  });
});
