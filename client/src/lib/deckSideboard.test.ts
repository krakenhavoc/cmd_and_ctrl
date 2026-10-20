// @vitest-environment jsdom
//
// deckSideboard.test.ts — #3833, #4054. The sideboard is read on the saved
// deck (the Decks page), from the printings route's rows, and not on the
// table.

import { describe, it, expect, afterEach, vi } from "vitest";

import DeckCard from "./components/decks/DeckCard.svelte";
import SideboardList from "./components/decks/SideboardList.svelte";
import { sideboardLines, sideboardTotal } from "./deckSideboard";
import type { DeckPrintingRow } from "./api";
import type { MyDeckInfo } from "./myDecks";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

const row = (name: string, count: number, board: string, entry = 0): DeckPrintingRow => ({
  entry,
  name,
  count,
  board,
  printing_id: "p",
  set: "s",
  collector_number: "1",
});

describe("sideboardLines", () => {
  it("keeps only the sideboard, one line per name, in name order", () => {
    const lines = sideboardLines([
      row("Kenrith, the Returned King", 1, "Commander"),
      row("Plains", 40, "Mainboard"),
      row("Living Wish", 1, "Sideboard"),
      row("Burning Wish", 2, "Sideboard"),
      row("Burning Wish", 1, "Sideboard"),
      row("Companion Card", 1, "Companion"),
    ]);
    expect(lines).toEqual([
      { name: "Burning Wish", count: 3 },
      { name: "Living Wish", count: 1 },
    ]);
    expect(sideboardTotal(lines)).toBe(4);
  });

  it("is empty for a deck with no sideboard", () => {
    expect(sideboardLines([row("Plains", 99, "Mainboard")])).toEqual([]);
    expect(sideboardTotal([])).toBe(0);
  });
});

const DECK: MyDeckInfo = {
  id: "d1",
  name: "Wish Tribal",
  commanders: ["Kenrith, the Returned King"],
  card_count: 100,
  updated_at: "2026-10-19T00:00:00Z",
};

function stubPrintings(rows: DeckPrintingRow[] | "fail") {
  const fetch = vi.fn(async (url: string) => {
    expect(url).toContain("/me/decks/d1/printings");
    if (rows === "fail") return new Response("nope", { status: 500 });
    return new Response(JSON.stringify({ revision: "r", cards: rows }), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  });
  vi.stubGlobal("fetch", fetch);
  return fetch;
}

async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("the saved deck's sideboard", () => {
  it("lists the sideboard read only, with its total", async () => {
    stubPrintings([
      row("Plains", 99, "Mainboard", 0),
      row("Burning Wish", 1, "Sideboard", 1),
      row("Living Wish", 2, "Sideboard", 2),
    ]);
    const { container } = render(
      SideboardList as never,
      { deck: DECK, onclose: () => {} } as never,
    );
    await settle();
    const section = container.querySelector('[data-testid="deck-sideboard"]')!;
    expect(section.querySelector("h4")?.textContent).toContain("(3)");
    expect(
      [...section.querySelectorAll("li")].map((li) => li.textContent?.replace(/\s+/g, " ").trim()),
    ).toEqual(["1× Burning Wish", "2× Living Wish"]);
    expect(section.textContent).not.toContain("Plains");
    // Read only: the only control is Close.
    expect([...section.querySelectorAll("button")].map((b) => b.textContent?.trim())).toEqual([
      "Close",
    ]);
  });

  it("says so when the deck has no sideboard", async () => {
    stubPrintings([row("Plains", 99, "Mainboard")]);
    const { container } = render(
      SideboardList as never,
      { deck: DECK, onclose: () => {} } as never,
    );
    await settle();
    expect(container.textContent).toContain("This deck has no sideboard.");
  });

  it("says when it could not load", async () => {
    stubPrintings("fail");
    const { container } = render(
      SideboardList as never,
      { deck: DECK, onclose: () => {} } as never,
    );
    await settle();
    expect(container.querySelector('[role="alert"]')).not.toBeNull();
  });

  it("opens from the deck card's Sideboard button, and closes again", async () => {
    stubPrintings([row("Burning Wish", 1, "Sideboard")]);
    const { container } = render(
      DeckCard as never,
      {
        deck: DECK,
        reportOpen: false,
        canRequest: false,
        requesting: false,
        requestLocked: false,
        busy: false,
        onReport: () => {},
        onPrintingSaved: () => {},
        onUpdated: () => {},
        onRequest: () => {},
        onRename: async () => "",
        onDelete: async () => {},
      } as never,
    );
    expect(container.querySelector('[data-testid="deck-sideboard"]')).toBeNull();
    const open = container.querySelector<HTMLButtonElement>(
      `button[aria-label="sideboard of Wish Tribal"]`,
    )!;
    expect(open.textContent?.trim()).toBe("Sideboard");
    click(open);
    await settle();
    expect(container.querySelector('[data-testid="deck-sideboard"]')?.textContent).toContain(
      "1× Burning Wish",
    );
    const close = [...container.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => b.textContent?.trim() === "Close",
    )!;
    click(close);
    await settle();
    expect(container.querySelector('[data-testid="deck-sideboard"]')).toBeNull();
  });
});
