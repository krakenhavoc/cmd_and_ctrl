// @vitest-environment jsdom
//
// deckUpdate.render.test.ts — #2036: the "Update deck" action on a saved deck.
//
//   - every saved deck offers it; the sheet has three ways in, "Re-pull from
//     ‹site›" first and only for a deck that has a link;
//   - Preview changes writes nothing and shows the diff and the new
//     coverage; Save commits the previewed list; Cancel closes without
//     writing;
//   - the request box appears only when the server can file one and there
//     are new cards to ask for, and its answer rides the deck's note;
//   - a Moxfield fetch error offers to paste the list instead.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Decks from "../routes/Decks.svelte";
import { L } from "./labels";
import type { DeckUpdateConfirm, DeckUpdatePreview } from "./deckUpdate";
import type { MyDeckInfo, MyDeckCoverage } from "./myDecks";
import { sessionFromOAuth, setSession } from "./session";
import { click, cleanup, flushSync, render } from "./test/render.svelte";

const COUNTS = { manual: 1, unreviewed: 0, caveats: 0, automated: 5, no_effect: 90 };
const COVERAGE: MyDeckCoverage = {
  counts: COUNTS,
  unknown: 0,
  as_printed: 95,
  resolved: 96,
  copies: COUNTS,
  unknown_copies: 0,
  as_printed_copies: 95,
  deck_size: 96,
};

function deck(over: Partial<MyDeckInfo>): MyDeckInfo {
  return {
    id: "d1",
    name: "Atraxa",
    commanders: ["Atraxa, Praetors' Voice"],
    card_count: 100,
    updated_at: "2026-09-19T08:00:00Z",
    coverage: COVERAGE,
    ...over,
  };
}

const PREVIEW: DeckUpdatePreview = {
  preview_id: "p-1",
  expires_at: "2099-01-01T00:00:00Z",
  source: "relink",
  source_url: "https://archidekt.com/decks/42",
  diff: {
    added: [{ name: "Doubling Season", count: 1 }],
    removed: [{ name: "Sol Ring", count: 1 }],
    changed: [{ name: "Forest", from: 10, to: 12 }],
    unchanged: 90,
    commanders_before: ["Atraxa, Praetors' Voice"],
    commanders_after: ["Atraxa, Praetors' Voice"],
    commander_changed: false,
  },
  no_change: false,
  printings_kept: 3,
  coverage_before: COVERAGE,
  coverage_after: { ...COVERAGE, counts: { ...COUNTS, manual: 2 }, as_printed_copies: 94 },
  report: {
    deck_name: "Linked",
    source: "archidekt",
    commanders: [],
    counts: { manual: 2, unreviewed: 0, caveats: 0, automated: 5, no_effect: 90 },
    copies: { manual: 2, unreviewed: 0, caveats: 0, automated: 5, no_effect: 90 },
    unknown_copies: 0,
    cards: [],
    unknown: [],
    violations: [],
  },
  request: { available: true, new_cards: ["Doubling Season"], already_requested: 1 },
  live_tables: 0,
};

interface Call {
  url: string;
  method: string;
  body?: string;
}

let calls: Call[];
let library: MyDeckInfo[];
let preview: DeckUpdatePreview;
let previewError: { status: number; body: unknown } | null;
let confirmRequest: DeckUpdateConfirm["request"];

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
  previewError = null;
  confirmRequest = undefined;
  preview = PREVIEW;
  sessionStorage.clear();
  library = [
    deck({}),
    deck({ id: "d2", name: "Linked", source_url: "https://archidekt.com/decks/42" }),
  ];
  setSession({
    ...sessionFromOAuth({ token: "tok", expiresAt: "2099-01-01T00:00:00Z", userID: "u1" }),
  });
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init?: RequestInit) => {
      const method = init?.method ?? "GET";
      calls.push({ url, method, body: init?.body as string | undefined });
      if (url === "/decks") return respond(200, { decks: [] });
      if (url === "/me/decks") return respond(200, { decks: library });
      if (url.endsWith("/update") && method === "POST") {
        if (previewError) return respond(previewError.status, previewError.body);
        return respond(200, preview);
      }
      if (url.endsWith("/update/confirm") && method === "POST") {
        const id = url.split("/")[3];
        const before = library.find((d) => d.id === id)!;
        const out: DeckUpdateConfirm = {
          deck: { ...before, card_count: 99, coverage: preview.coverage_after },
          diff: preview.diff,
          request: confirmRequest,
        };
        return respond(200, out);
      }
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

const button = (c: ParentNode, text: string): HTMLButtonElement | undefined =>
  [...c.querySelectorAll<HTMLButtonElement>("button")].find((b) => b.textContent?.trim() === text);

function type(el: HTMLInputElement | HTMLTextAreaElement, value: string): void {
  el.value = value;
  el.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

const card = (c: HTMLElement, name: string): HTMLElement =>
  c.querySelector<HTMLElement>(`article[aria-label="${name}"]`)!;

const sheet = (c: HTMLElement): HTMLElement | null =>
  c.querySelector<HTMLElement>(`section[aria-label="${L.deckUpdateSheet}"]`);

const posts = (suffix: string): Call[] =>
  calls.filter((c) => c.url.endsWith(suffix) && c.method === "POST");

async function open(c: HTMLElement, name: string): Promise<HTMLElement> {
  click(button(card(c, name), L.updateDeck)!);
  await settle();
  return sheet(c)!;
}

describe("Update deck on a saved deck (#2036)", () => {
  it("is offered on every saved deck", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    expect(button(card(container, "Atraxa"), L.updateDeck)).toBeDefined();
    expect(button(card(container, "Linked"), L.updateDeck)).toBeDefined();
    expect(sheet(container)).toBeNull();
  });

  it("offers re-pull first for a linked deck, and only new link and paste for a pasted one", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();

    let s = await open(container, "Linked");
    const tabs = [
      ...s.querySelectorAll<HTMLButtonElement>(`[aria-label="${L.deckUpdateChoices}"] button`),
    ];
    expect(tabs.map((t) => t.textContent?.trim())).toEqual([
      "Re-pull from archidekt.com",
      "New link",
      "Paste list",
    ]);
    expect(tabs[0].getAttribute("aria-pressed")).toBe("true");
    click(button(s, L.cancelUpdate)!);
    expect(sheet(container)).toBeNull();

    s = await open(container, "Atraxa");
    const plain = [
      ...s.querySelectorAll<HTMLButtonElement>(`[aria-label="${L.deckUpdateChoices}"] button`),
    ];
    expect(plain.map((t) => t.textContent?.trim())).toEqual(["New link", "Paste list"]);
    expect(plain[1].getAttribute("aria-pressed")).toBe("true");
  });

  it("previews a re-pull without writing, shows the diff, then saves what was previewed", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    const s = await open(container, "Linked");

    click(button(s, L.previewUpdate)!);
    await settle();
    expect(JSON.parse(posts("/decks/d2/update")[0].body!)).toEqual({ source: "relink" });
    expect(posts("/update/confirm")).toHaveLength(0);

    const text = (s.textContent ?? "").replace(/\s+/g, " ");
    expect(text).toContain("1 added, 1 removed, 1 changed");
    expect(text).toContain("Doubling Season");
    expect(text).toContain("Sol Ring");
    expect(text).toContain("Forest 10 → 12");
    expect(text).toContain("Now: 95 of 96 cards play as printed");
    expect(text).toContain("After: 94 of 96 cards play as printed");
    expect(text).toContain("3 card rows keep the art you chose");

    // The box to request the new card is on by default.
    const box = s.querySelector<HTMLInputElement>('input[type="checkbox"]')!;
    expect(box.checked).toBe(true);
    expect(s.textContent).toContain("Request the 1 new missing card (1 already requested)");

    confirmRequest = {
      status: "joined",
      issue_url: "https://github.com/krakenhavoc/cmd_and_ctrl/issues/1700",
      issue_number: 1700,
      new_cards: ["Doubling Season"],
    };
    click(button(s, L.saveUpdate)!);
    await settle();
    expect(JSON.parse(posts("/update/confirm")[0].body!)).toEqual({
      preview_id: "p-1",
      request_missing: true,
    });
    // The sheet is gone, the deck is the updated one, and its note says what happened.
    expect(sheet(container)).toBeNull();
    expect(card(container, "Linked").textContent).toContain("99 cards");
    const note = card(container, "Linked").querySelector(".outcome")!;
    expect(note.textContent).toContain("Updated Linked: 1 added, 1 removed, 1 changed.");
    expect(note.textContent).toContain("Added 1 new card to the deck's open issue.");
    expect(note.querySelector("a")?.getAttribute("href")).toBe(
      "https://github.com/krakenhavoc/cmd_and_ctrl/issues/1700",
    );
  });

  it("saves without requesting when the box is unticked", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    const s = await open(container, "Linked");
    click(button(s, L.previewUpdate)!);
    await settle();
    click(s.querySelector<HTMLInputElement>('input[type="checkbox"]')!);
    click(button(s, L.saveUpdate)!);
    await settle();
    expect(JSON.parse(posts("/update/confirm")[0].body!).request_missing).toBe(false);
    expect(card(container, "Linked").textContent).not.toContain("issue");
  });

  it("explains why it cannot request, and offers no box", async () => {
    preview = {
      ...PREVIEW,
      request: {
        available: false,
        reason: "Sign in with Discord to request cards.",
        new_cards: ["Doubling Season"],
        already_requested: 0,
      },
    };
    const { container } = render(Decks as never, {} as never);
    await settle();
    const s = await open(container, "Linked");
    click(button(s, L.previewUpdate)!);
    await settle();
    expect(s.querySelector('input[type="checkbox"]')).toBeNull();
    expect(s.textContent).toContain("Sign in with Discord to request cards.");
    click(button(s, L.saveUpdate)!);
    await settle();
    expect(JSON.parse(posts("/update/confirm")[0].body!).request_missing).toBe(false);
  });

  it("sends a pasted list, and warns that it replaces the link", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    const s = await open(container, "Linked");
    click(button(s, "Paste list")!);
    expect(s.textContent).toContain("replaces this deck's link");
    expect(button(s, L.previewUpdate)!.disabled).toBe(true);
    type(s.querySelector<HTMLTextAreaElement>("textarea")!, "1 Sol Ring\n99 Forest");
    expect(button(s, L.previewUpdate)!.disabled).toBe(false);
    click(button(s, L.previewUpdate)!);
    await settle();
    expect(JSON.parse(posts("/update")[0].body!)).toEqual({ list: "1 Sol Ring\n99 Forest" });
  });

  it("sends a new link", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    const s = await open(container, "Atraxa");
    click(button(s, "New link")!);
    type(
      s.querySelector<HTMLInputElement>('input[type="url"]')!,
      " https://archidekt.com/decks/9 ",
    );
    click(button(s, L.previewUpdate)!);
    await settle();
    expect(JSON.parse(posts("/update")[0].body!)).toEqual({ url: "https://archidekt.com/decks/9" });
  });

  it("cancelling after a preview writes nothing", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    const s = await open(container, "Linked");
    click(button(s, L.previewUpdate)!);
    await settle();
    click(button(s, L.cancelUpdate)!);
    expect(sheet(container)).toBeNull();
    expect(posts("/update/confirm")).toHaveLength(0);
    expect(card(container, "Linked").textContent).toContain("100 cards");
  });

  it("shows the server's refusal and offers to paste when Moxfield blocks the fetch", async () => {
    previewError = {
      status: 424,
      body: { error: "Moxfield blocks our server.", hint: "paste_list" },
    };
    const { container } = render(Decks as never, {} as never);
    await settle();
    const s = await open(container, "Linked");
    click(button(s, L.previewUpdate)!);
    await settle();
    expect(s.querySelector("[role=alert]")?.textContent).toContain("Moxfield blocks our server.");
    click(button(s, "Paste the list instead")!);
    expect(s.querySelector("textarea")).not.toBeNull();
    expect(s.querySelector("[role=alert]")).toBeNull();
  });

  it("goes back to the start when the preview has expired", async () => {
    const { container } = render(Decks as never, {} as never);
    await settle();
    const s = await open(container, "Linked");
    click(button(s, L.previewUpdate)!);
    await settle();
    const real = globalThis.fetch;
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init?: RequestInit) =>
        url.endsWith("/update/confirm")
          ? respond(410, {
              error: "that preview expired or was already used; preview the update again",
            })
          : real(url, init),
      ),
    );
    click(button(s, L.saveUpdate)!);
    await settle();
    expect(s.querySelector("[role=alert]")?.textContent).toContain("preview the update again");
    expect(button(s, L.previewUpdate)).toBeDefined();
    expect(button(s, L.saveUpdate)).toBeUndefined();
  });

  it("a commander change is called out", async () => {
    preview = {
      ...PREVIEW,
      diff: {
        ...PREVIEW.diff,
        commander_changed: true,
        commanders_after: ["Tymna the Weaver", "Thrasios, Triton Hero"],
      },
    };
    const { container } = render(Decks as never, {} as never);
    await settle();
    const s = await open(container, "Linked");
    click(button(s, L.previewUpdate)!);
    await settle();
    expect(s.textContent).toContain(
      "Commander: Atraxa, Praetors' Voice → Tymna the Weaver / Thrasios, Triton Hero",
    );
  });
});
