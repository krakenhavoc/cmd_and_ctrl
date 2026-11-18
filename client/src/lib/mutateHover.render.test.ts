// @vitest-environment jsdom
//
// ADR 0223: the hover panel says what a mutated permanent is made of,
// top card first.

import { describe, it, expect, beforeEach, afterEach } from "vitest";

import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import { hoveredCard } from "./cardTypes";
import type { CardView, GameView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const view: GameView = {
  id: "g",
  state: "active",
  seats: [],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
} as unknown as GameView;

let realFetch: unknown;

beforeEach(() => {
  hoveredCard.set(null);
  const g = globalThis as Record<string, unknown>;
  realFetch = g.fetch;
  g.fetch = () =>
    Promise.resolve({
      ok: true,
      json: () => Promise.resolve({ name: "Everquill Phoenix", type_line: "", oracle_text: "" }),
    } as unknown as Response);
});

afterEach(() => {
  cleanup();
  (globalThis as Record<string, unknown>).fetch = realFetch;
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    flushSync();
  }
}

describe("mutate in the hover panel", () => {
  it("names a mutated permanent's cards, top first", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set({
      instance_id: "elves",
      name: "Everquill Phoenix",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: "phx",
      type_line: "Creature — Phoenix",
      merged_from: [
        { name: "Everquill Phoenix", image: "/cards/phx/image" },
        { name: "Llanowar Elves", image: "/cards/elf/image" },
      ],
    } as unknown as CardView);
    await settle();
    expect(container.querySelector(".info-foot")?.textContent).toContain(
      "mutated: Everquill Phoenix over Llanowar Elves",
    );
  });

  it("says nothing about mutating on an ordinary creature", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set({
      instance_id: "bear",
      name: "Grizzly Bears",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: "bear",
      type_line: "Creature — Bear",
    } as unknown as CardView);
    await settle();
    expect(container.textContent ?? "").not.toContain("mutated:");
  });
});
