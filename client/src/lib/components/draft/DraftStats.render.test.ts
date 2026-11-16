// @vitest-environment jsdom
//
// #5848 §1, §2, §3 and §6: Deck view's cards drag and preview like the
// pick screen's, and the box of reasons under the large preview.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { DraftCard } from "../../draft";
import { hidePreview } from "../../draftPreview.svelte";
import { settings } from "../../settings";
import { cleanup, click, flushSync, render } from "../../test/render.svelte";
import DraftDeckColumns from "./DraftDeckColumns.svelte";
import DraftReasons from "./DraftReasons.svelte";

const card = (id: string, name: string, over: Partial<DraftCard> = {}): DraftCard => ({
  id,
  name,
  mana_value: 2,
  colors: ["U"],
  type_line: "Creature",
  ...over,
});
const picks = [card("c2", "Thundertrap Trainer"), card("c3", "Mindwhisker")];

function fire(target: Element | Window, type: string, x = 0, y = 0): void {
  const ev = new MouseEvent(type, { bubbles: true, cancelable: true, clientX: x, clientY: y });
  Object.defineProperty(ev, "pointerId", { value: 1 });
  Object.defineProperty(ev, "pointerType", { value: "mouse" });
  target.dispatchEvent(ev);
  flushSync();
}

let hit: Element | null = null;
beforeEach(() => {
  vi.useFakeTimers();
  document.elementFromPoint = () => hit;
  hit = null;
  settings.update((s) => ({ ...s, display: { ...s.display, hoverDelayMs: 300 } }));
});
afterEach(() => {
  hidePreview();
  cleanup();
  vi.runOnlyPendingTimers();
  vi.useRealTimers();
  vi.restoreAllMocks();
});

const text = (el: Element | null): string => (el?.textContent ?? "").replace(/\s+/g, " ").trim();
const zone = (c: HTMLElement, id: "main" | "side"): HTMLElement =>
  c.querySelector<HTMLElement>(`[data-testid="draft-deckview-${id}"]`)!;
const tile = (c: HTMLElement, id: string): HTMLElement =>
  c.querySelector<HTMLElement>(`[data-move-id="${id}"]`)!;

function deck(side: string[] = [], onmove = vi.fn()) {
  return { ...render(DraftDeckColumns as never, { cards: picks, side, onmove } as never), onmove };
}

describe("Deck view's cards (#5848 §1, §2)", () => {
  it("are drop zones for Main and the Sideboard", () => {
    const { container: c } = deck();
    expect(zone(c, "main").getAttribute("data-drop-zone")).toBe("main");
    expect(zone(c, "side").getAttribute("data-drop-zone")).toBe("side");
  });

  it("drag from Main onto the Sideboard with the pointer, highlighting the target", () => {
    const { container: c, onmove } = deck();
    const btn = tile(c, "c2");
    fire(btn, "pointerdown", 10, 10);
    hit = zone(c, "side");
    fire(window, "pointermove", 60, 200);
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).not.toBeNull();
    expect(btn.style.opacity).toBe("0.35");
    expect(zone(c, "side").classList.contains("drop")).toBe(true);
    fire(window, "pointerup", 60, 200);
    expect(onmove).toHaveBeenCalledExactlyOnceWith("c2", "side");
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).toBeNull();
    expect(text(c.querySelector('[data-testid="draft-deckview-said"]'))).toBe(
      "Moved Thundertrap Trainer to sideboard.",
    );
  });

  it("drag from the Sideboard back to Main, and a drop where it was is no move", () => {
    const { container: c, onmove } = deck(["c3"]);
    fire(tile(c, "c3"), "pointerdown", 10, 10);
    hit = zone(c, "side");
    fire(window, "pointermove", 60, 200);
    fire(window, "pointerup", 60, 200);
    expect(onmove).not.toHaveBeenCalled();

    fire(tile(c, "c3"), "pointerdown", 10, 10);
    hit = zone(c, "main");
    fire(window, "pointermove", 60, 200);
    fire(window, "pointerup", 60, 200);
    expect(onmove).toHaveBeenCalledExactlyOnceWith("c3", "main");
  });

  it("still move with a click, and say so", () => {
    const { container: c, onmove } = deck();
    click(tile(c, "c2"));
    expect(onmove).toHaveBeenCalledExactlyOnceWith("c2", "side");
    expect(tile(c, "c2").getAttribute("aria-label")).toContain("Move to sideboard: ");
    expect(text(c.querySelector('[data-testid="draft-deckview-said"]'))).toBe(
      "Moved Thundertrap Trainer to sideboard.",
    );
  });

  it("show the large preview after the delay, and not while dragging", () => {
    const { container: c } = deck();
    const pv = (): HTMLElement | null => document.querySelector('[data-testid="draft-preview"]');
    fire(tile(c, "c2"), "pointerenter", 5, 5);
    vi.advanceTimersByTime(310);
    flushSync();
    expect(pv()?.getAttribute("data-card-id")).toBe("c2");
    fire(tile(c, "c2"), "pointerdown", 10, 10);
    hit = zone(c, "side");
    fire(window, "pointermove", 60, 200);
    expect(pv()).toBeNull();
    fire(window, "pointerup", 60, 200);
  });
});

describe("the box of reasons under the preview (#5848 §3)", () => {
  const rated = card("r1", "Doom Blade", {
    colors: ["B"],
    mana_value: 2,
    type_line: "Instant",
    oracle_text: "Destroy target nonblack creature.",
    rating: {
      source: "17lands",
      value: 60.4,
      tier: "A-",
      fetched: "2026-11-11",
      avg_seen: 2.2,
    },
  });

  it("says the rating, a few lines, and the numbers with their source", () => {
    const { container: c } = render(DraftReasons as never, { card: rated, setAvg: 57.2 } as never);
    expect(text(c.querySelector("h3"))).toBe("Rating 60.4 (A-)");
    const lines = [...c.querySelectorAll("li")].map((l) => text(l));
    expect(lines).toEqual([
      "Wins 60.4% of games when drawn, +3.2 against this set's average.",
      "Removal: it deals with a creature, so it fits almost any deck in its colour.",
    ]);
    expect(text(c.querySelector(".foot"))).toBe(
      "GIH WR 60.4% Mana value 2 Source: 17Lands, 2026-11-11",
    );
    expect(c.textContent).not.toMatch(/ALSA|gone by pick/);
  });

  it("draws nothing for a card with no rating", () => {
    const { container: c } = render(
      DraftReasons as never,
      {
        card: card("r2", "Island"),
      } as never,
    );
    expect(c.querySelector('[data-testid="draft-reasons"]')).toBeNull();
  });
});
