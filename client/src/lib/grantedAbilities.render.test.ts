// @vitest-environment jsdom
//
// ADR 0093 Decision 8: the hover panel lists the abilities other
// permanents granted this one, one row per grantor, beside the oracle
// text the card never printed them in.

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
      json: () => Promise.resolve({ name: "Bear", type_line: "Creature — Bear", oracle_text: "" }),
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

describe("granted abilities in the hover panel", () => {
  it("lists each granted ability with its grantor, one row per grantor", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set({
      instance_id: "bear",
      name: "Bear",
      owner: "me",
      controller: "me",
      known_by_you: true,
      type_line: "Creature — Bear",
      granted_abilities: [
        {
          text: "{T}: Add one mana of any color.",
          source_id: "r1",
          source_name: "Cryptolith Rite",
        },
        {
          text: "{T}: Add one mana of any color.",
          source_id: "r2",
          source_name: "Cryptolith Rite",
        },
        { text: "When this creature dies, draw a card." },
      ],
    } as unknown as CardView);
    await settle();
    const rows = container.querySelectorAll("ul.granted li");
    expect(rows).toHaveLength(3);
    expect(rows[0].textContent).toContain("{T}: Add one mana of any color.");
    expect(rows[0].textContent).toContain("from Cryptolith Rite");
    expect(rows[2].textContent).not.toContain("from");
  });

  it("renders no list for a permanent with nothing granted", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set({
      instance_id: "bear",
      name: "Bear",
      owner: "me",
      controller: "me",
      known_by_you: true,
      type_line: "Creature — Bear",
    } as unknown as CardView);
    await settle();
    expect(container.querySelector("ul.granted")).toBeNull();
  });
});
