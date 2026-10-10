// @vitest-environment jsdom
//
// ADR 0145: the hover panel shows a melded permanent's two cards and a
// meld card's combined back face, and says what a melded permanent is
// made of.

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
      json: () => Promise.resolve({ name: "Urza, Planeswalker", type_line: "", oracle_text: "" }),
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

describe("meld in the hover panel", () => {
  it("shows a melded permanent's two cards and names them", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set({
      instance_id: "urza",
      name: "Urza, Planeswalker",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: "pw",
      type_line: "Legendary Planeswalker — Urza",
      melded_from: [
        { name: "Urza, Lord Protector", image: "/cards/a/image" },
        { name: "The Mightstone and Weakstone", image: "/cards/b/image" },
      ],
    } as unknown as CardView);
    await settle();
    const insets = container.querySelectorAll(".meld-inset img");
    expect(insets).toHaveLength(2);
    expect(insets[0].getAttribute("src")).toBe("/cards/a/image");
    expect(container.querySelector(".info-foot")?.textContent).toContain(
      "melded from Urza, Lord Protector + The Mightstone and Weakstone",
    );
  });

  it("previews the permanent a meld card melds into", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set({
      instance_id: "lp",
      name: "Urza, Lord Protector",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: "lp",
      type_line: "Legendary Creature — Human Artificer",
      melds_into: { name: "Urza, Planeswalker", image: "/cards/pw/image" },
    } as unknown as CardView);
    await settle();
    const insets = container.querySelectorAll(".meld-inset");
    expect(insets).toHaveLength(1);
    expect(insets[0].getAttribute("title")).toBe("Melds into Urza, Planeswalker");
    expect(container.querySelector(".info-foot")).toBeNull();
  });
});
