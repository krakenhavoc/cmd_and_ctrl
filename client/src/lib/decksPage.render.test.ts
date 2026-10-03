// @vitest-environment jsdom
//
// decksPage.render.test.ts — ADR 0112 §3, the one decks page (#/decks):
//
//   - a check shows "N of M cards play as printed", and writes nothing
//     (owner answer 2): saving is always the explicit button;
//   - "Request missing cards" and "Save to my decks" are separate, and
//     each is replaced by the right prompt for a visitor who may not use
//     it (§3 item 3);
//   - the library keeps rename, delete and the report, and every deck,
//     pasted or linked, asks for its missing cards by deck_id (item 5);
//   - the pre-built decks are a read-only list (owner answer 4);
//   - signing in from the page stores the way back (item 7).
//
// The lobby's saved-deck picker shares the coverage line, so its test
// stays here too.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Decks from "../routes/Decks.svelte";
import YourDecksPicker from "./components/YourDecksPicker.svelte";
import type { CoverageReport } from "./deckcheck";
import { AFTER_SIGN_IN_KEY, AFTER_SIGN_IN_TEXT_KEY, takeAfterSignIn } from "./decksPage";
import type { MyDeckInfo } from "./myDecks";
import type { PrebuiltDeck } from "./prebuiltDecks";
import { sessionFromOAuth, setSession, type Session } from "./session";
import { click, cleanup, flushSync, render } from "./test/render.svelte";

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

const CHECKED: CoverageReport = {
  deck_name: "Weekend deck",
  source: "archidekt",
  source_url: "https://archidekt.com/decks/42",
  deck_key: "archidekt:42",
  commanders: ["Atraxa, Praetors' Voice"],
  counts: { manual: 1, unreviewed: 1, caveats: 1, automated: 6, no_effect: 1 },
  cards: [
    { name: "Doubling Season", oracle_id: "o1", count: 1, bucket: "manual" },
    { name: "Sol Ring", oracle_id: "o2", count: 1, bucket: "automated" },
  ],
  unknown: [],
  violations: [],
};

const PREBUILT: PrebuiltDeck[] = [
  {
    id: "izzet-aggro",
    name: "Raid and Ransack",
    archetype: "aggro",
    commander: "Mary Read and Anne Bonny",
    card_count: 100,
    coverage: { cards: 60, full: 58, caveats: 2, unreviewed: 0, basics: 2, unregistered: 0 },
  },
];

interface Call {
  url: string;
  method: string;
  body?: string;
}

let calls: Call[];
let library: MyDeckInfo[];
let renameStatus: number;
let saveStatus: number;

function respond(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: "",
    json: async () => body,
    clone: () => ({ json: async () => body }),
  } as unknown as Response;
}

function signIn(over: Partial<Session> = {}): void {
  setSession({
    ...sessionFromOAuth({ token: "tok", expiresAt: "2099-01-01T00:00:00Z", userID: "u1" }),
    ...over,
  });
}

beforeEach(() => {
  calls = [];
  renameStatus = 200;
  saveStatus = 201;
  sessionStorage.clear();
  library = [
    deck({}),
    deck({
      id: "d2",
      name: "Linked",
      source_url: "https://moxfield.com/decks/abc",
      coverage: { counts: COUNTS, unknown: 0, as_printed: 3, resolved: 4 },
    }),
  ];
  signIn();
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init?: RequestInit) => {
      const method = init?.method ?? "GET";
      calls.push({ url, method, body: init?.body as string | undefined });
      if (url === "/deck-coverage") return respond(200, CHECKED);
      if (url === "/deck-requests") {
        return respond(201, {
          status: "filed",
          issue_url: "https://github.com/krakenhavoc/cmd_and_ctrl/issues/1700",
          issue_number: 1700,
        });
      }
      if (url === "/decks") return respond(200, { decks: PREBUILT });
      if (url === "/me/decks" && method === "POST") {
        if (saveStatus === 409) {
          return respond(409, {
            error: "Your deck library is full (200 decks). Delete one below to save this one.",
            code: "library_full",
          });
        }
        const body = JSON.parse(init?.body as string) as { name?: string };
        const replaced = library.find((d) => d.name === body.name);
        return respond(replaced ? 200 : 201, {
          deck: deck({ id: replaced?.id ?? "d9", name: body.name ?? "Untitled deck" }),
          replaced: Boolean(replaced),
        });
      }
      if (url === "/me/decks") return respond(200, { decks: library });
      if (url.endsWith("/coverage")) {
        return respond(200, { ...CHECKED, source: "text", source_url: undefined });
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
  sessionStorage.clear();
});

async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

const button = (c: HTMLElement, text: string): HTMLButtonElement | undefined =>
  [...c.querySelectorAll<HTMLButtonElement>("button")].find((b) => b.textContent?.trim() === text);

function type(el: HTMLInputElement | HTMLTextAreaElement, value: string): void {
  el.value = value;
  el.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

async function checkLink(c: HTMLElement): Promise<void> {
  type(c.querySelector<HTMLInputElement>('input[aria-label="deck link"]')!, CHECKED.source_url!);
  click(button(c, "Check this deck")!);
  await settle();
}

function posts(url: string): Call[] {
  return calls.filter((c) => c.url === url && c.method === "POST");
}

describe("#/decks: the check (§3 item 1)", () => {
  it("shows the report as 'N of M cards play as printed', and saves nothing", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    await checkLink(container);
    const text = container.textContent ?? "";
    expect(text).toContain("Weekend deck");
    expect(text).toContain("7 of 10 cards play as printed");
    expect(text).toContain("Doubling Season");
    // Owner answer 2: a check writes nothing.
    expect(posts("/me/decks")).toHaveLength(0);
    expect(posts("/deck-requests")).toHaveLength(0);
  });

  it("offers Request missing cards and Save to my decks as two buttons", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    await checkLink(container);
    expect(button(container, "Request missing cards")).toBeDefined();
    expect(button(container, "Save to my decks")).toBeDefined();
    const name = container.querySelector<HTMLInputElement>(
      'input[aria-label="name to save this deck as"]',
    )!;
    expect(name.value).toBe("Weekend deck");
  });

  it("requests without saving", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    await checkLink(container);
    click(button(container, "Request missing cards")!);
    await settle();
    expect(JSON.parse(posts("/deck-requests")[0].body!)).toEqual({ url: CHECKED.source_url });
    expect(posts("/me/decks")).toHaveLength(0);
    expect(container.textContent).toContain("Filed a new issue for the missing cards.");
  });

  it("saves with the name in the field, without requesting, and lists the deck", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    await checkLink(container);
    type(
      container.querySelector<HTMLInputElement>('input[aria-label="name to save this deck as"]')!,
      "Friday night",
    );
    click(button(container, "Save to my decks")!);
    await settle();
    expect(JSON.parse(posts("/me/decks")[0].body!)).toEqual({
      url: CHECKED.source_url,
      name: "Friday night",
    });
    expect(posts("/deck-requests")).toHaveLength(0);
    expect(container.textContent).toContain("Saved Friday night to your decks.");
    const names = [...container.querySelectorAll('ul[aria-label="your decks"] .name')].map(
      (n) => n.textContent,
    );
    expect(names).toEqual(["Friday night", "Atraxa", "Linked"]);
  });

  it("reads Replace ‹name› before sending when the name is already saved", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    await checkLink(container);
    type(
      container.querySelector<HTMLInputElement>('input[aria-label="name to save this deck as"]')!,
      "Atraxa",
    );
    expect(button(container, "Replace Atraxa")).toBeDefined();
    click(button(container, "Replace Atraxa")!);
    await settle();
    expect(container.textContent).toContain("Replaced Atraxa in your decks.");
  });

  it("keeps the report on screen when the library is full", async () => {
    saveStatus = 409;
    const { container } = render(Decks as never, {} as never);
    await settle();
    await checkLink(container);
    click(button(container, "Save to my decks")!);
    await settle();
    expect(container.textContent).toContain("Your deck library is full (200 decks).");
    expect(container.textContent).toContain("7 of 10 cards play as printed");
  });
});

describe("#/decks: who sees what (§3 item 3)", () => {
  it("lets a signed-out visitor check, and asks them to sign in for the rest", async () => {
    setSession(null);
    const { container } = render(Decks as never, {} as never);
    await settle();
    // Nothing session-gated is asked for.
    expect(calls).toHaveLength(0);
    expect(container.textContent).toContain("Sign in with Discord to see them.");
    await checkLink(container);
    const text = container.textContent ?? "";
    expect(text).toContain("7 of 10 cards play as printed");
    expect(text).toContain("Sign in with Discord to request these cards");
    expect(text).toContain("Sign in with Discord to save this deck");
    expect(button(container, "Request missing cards")).toBeUndefined();
    expect(button(container, "Save to my decks")).toBeUndefined();
    expect(container.querySelector('ul[aria-label="your decks"]')).toBeNull();
    expect(container.textContent).not.toContain("Pre-built decks");
    expect(calls.map((c) => c.url)).toEqual(["/deck-coverage"]);
  });

  it("stores the way back before a signed-out visitor signs in (§3 item 7)", async () => {
    setSession(null);
    const { container } = render(Decks as never, {} as never);
    await settle();
    click(container.querySelector<HTMLButtonElement>('button[role="tab"]:nth-child(2)')!);
    type(container.querySelector<HTMLTextAreaElement>('textarea[aria-label="decklist"]')!, "1 Sol");
    click(button(container, "Check this deck")!);
    await settle();
    const signInLink = [...container.querySelectorAll<HTMLAnchorElement>("a")].find((a) =>
      a.textContent?.includes("Sign in with Discord to save this deck"),
    )!;
    signInLink.addEventListener("click", (e) => e.preventDefault());
    click(signInLink as unknown as HTMLButtonElement);
    expect(sessionStorage.getItem(AFTER_SIGN_IN_KEY)).toBe("#/decks");
    expect(sessionStorage.getItem(AFTER_SIGN_IN_TEXT_KEY)).toBe("1 Sol");
  });

  it("checks the pasted list again after the sign-in brings the visitor back", async () => {
    sessionStorage.setItem(AFTER_SIGN_IN_KEY, "#/decks");
    sessionStorage.setItem(AFTER_SIGN_IN_TEXT_KEY, "1 Sol Ring");
    expect(takeAfterSignIn()).toBe("#/decks");
    const { container } = render(Decks as never, {} as never);
    await settle();
    expect(JSON.parse(posts("/deck-coverage")[0].body!)).toEqual({ text: "1 Sol Ring" });
    expect(container.textContent).toContain("7 of 10 cards play as printed");
    expect(button(container, "Save to my decks")).toBeDefined();
  });

  it("tells a guest seat to link Discord at their table, never to sign in", async () => {
    setSession({
      token: "guest",
      expiresAt: "2099-01-01T00:00:00Z",
      principal: {
        role: "player",
        user_id: "00000000-0000-0000-0000-000000000000",
        name: "Guest",
        game_id: "g1",
        player_id: "p1",
        issued_at: "2026-01-01T00:00:00Z",
        expires_at: "2099-01-01T00:00:00Z",
      },
      gameID: "g1",
      playerID: "p1",
    });
    const { container } = render(Decks as never, {} as never);
    await settle();
    await checkLink(container);
    const text = container.textContent ?? "";
    expect(text).toContain("Link Discord from your table's menu to request these cards");
    expect(text).toContain("Link Discord from your table's menu to save this deck");
    expect(text).not.toContain("Sign in with Discord");
    expect(calls.some((c) => c.url === "/me/decks")).toBe(false);
  });

  it("hides request, save and the library from the admin token", async () => {
    setSession({
      token: "admin",
      expiresAt: "2099-01-01T00:00:00Z",
      principal: {
        role: "admin",
        name: "admin",
        issued_at: "2026-01-01T00:00:00Z",
        expires_at: "2099-01-01T00:00:00Z",
      },
    });
    const { container } = render(Decks as never, {} as never);
    await settle();
    await checkLink(container);
    const text = container.textContent ?? "";
    expect(text).toContain("7 of 10 cards play as printed");
    expect(text).not.toContain("Request missing cards");
    expect(text).not.toContain("Save to my decks");
    expect(text).not.toContain("Sign in with Discord");
    expect(text).not.toContain("Your decks");
    expect(calls.some((c) => c.url === "/me/decks")).toBe(false);
  });
});

describe("#/decks: your decks (§3 items 1.4, 5 and 9)", () => {
  it("says past decks are not recovered when the library is empty", async () => {
    library = [];
    const { container } = render(Decks as never, {} as never);
    await settle();
    expect(container.textContent).toContain(
      "No saved decks yet. Decks you save here, or import at a table while signed in, are kept.",
    );
  });

  it("lists each saved deck with its playable count and source link", async () => {
    const { container } = render(Decks as never, {} as never);
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

  it("requests a pasted deck's missing cards by deck_id, as well as a link deck's", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    click(
      container.querySelector<HTMLButtonElement>(
        'button[aria-label="request missing cards for Atraxa"]',
      )!,
    );
    await settle();
    click(
      container.querySelector<HTMLButtonElement>(
        'button[aria-label="request missing cards for Linked"]',
      )!,
    );
    await settle();
    expect(posts("/deck-requests").map((c) => JSON.parse(c.body!))).toEqual([
      { deck_id: "d1" },
      { deck_id: "d2" },
    ]);
    expect(container.textContent).toContain("Filed a new issue for the missing cards.");
  });

  it("opens the full report in place", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    click(button(container, "Coverage report")!);
    await settle();
    expect(calls.some((c) => c.url === "/me/decks/d1/coverage")).toBe(true);
    expect(container.textContent).toContain("Doubling Season");
  });

  it("renames, and shows the server's message on a conflict", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    click(button(container, "Rename")!);
    flushSync();
    type(container.querySelector<HTMLInputElement>("form.rename input")!, "Atraxa Prime");
    click(button(container, "Save")!);
    await settle();
    const patch = calls.find((c) => c.method === "PATCH")!;
    expect(patch.url).toBe("/me/decks/d1");
    expect(JSON.parse(patch.body!)).toEqual({ name: "Atraxa Prime" });
    expect(container.textContent).toContain("Atraxa Prime");

    renameStatus = 409;
    click(button(container, "Rename")!);
    flushSync();
    type(container.querySelector<HTMLInputElement>("form.rename input")!, "Linked");
    click(button(container, "Save")!);
    await settle();
    expect(container.textContent).toContain("you already have a deck with that name");
  });

  it("deletes only after a confirm", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    click(container.querySelector<HTMLButtonElement>('button[aria-label="delete Atraxa"]')!);
    flushSync();
    expect(calls.some((c) => c.method === "DELETE")).toBe(false);
    expect(container.textContent).toContain("Delete Atraxa?");
    click(button(container, "Delete")!);
    await settle();
    expect(calls.find((c) => c.method === "DELETE")?.url).toBe("/me/decks/d1");
    expect(container.textContent).not.toContain("7 of 10 cards play as printed");
    expect(container.textContent).toContain("3 of 4 cards play as printed");
  });
});

describe("#/decks: pre-built decks (owner answer 4)", () => {
  it("lists them read-only, with each one's as-printed line", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    const list = container.querySelector<HTMLElement>('ul[aria-label="pre-built decks"]')!;
    expect(list.textContent).toContain("Raid and Ransack");
    expect(list.textContent).toContain("Every card is implemented · 58 of 60 nonbasic cards");
    // Read-only: no control of any kind, so nothing can copy it.
    expect(list.querySelectorAll("button, input, form")).toHaveLength(0);
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
