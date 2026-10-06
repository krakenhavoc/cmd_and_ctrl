// @vitest-environment jsdom
//
// #2257, the rendered half: the damage badge on an indestructible
// creature carries the indestructible shield and says why the damage
// didn't destroy it, and the hover panel says it in words.
// damageBadge is unit-tested in damageBadge.test.ts.

import { describe, it, expect, beforeEach, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import { hoveredCard } from "./cardTypes";
import type { CardView, GameView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const solphim = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "solphim",
    name: "Solphim, Mayhem Dominus",
    owner: "bot",
    controller: "bot",
    known_by_you: true,
    type_line: "Legendary Creature — Phyrexian Horror",
    power: 5,
    toughness: 4,
    damage_marked: 4,
    ...extra,
  }) as unknown as CardView;

const indestructible = { abilities: ["indestructible"], counters: { indestructible: 1 } };

const view: GameView = {
  id: "g",
  state: "active",
  seats: [],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: { number: 9, active_seat: 2, priority_holder: 2, phase: "combat", step: "combat_damage" },
} as unknown as GameView;

let realFetch: unknown;

beforeEach(() => {
  hoveredCard.set(null);
  const g = globalThis as Record<string, unknown>;
  realFetch = g.fetch;
  g.fetch = () =>
    Promise.resolve({
      ok: true,
      json: () => Promise.resolve({ name: "Solphim, Mayhem Dominus", oracle_text: "" }),
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

describe("the damage badge", () => {
  it("shows the indestructible shield and explains lethal damage it survived", () => {
    const { container } = render(Card as never, { card: solphim(indestructible) } as never);
    const badge = container.querySelector(".badge.damage");
    expect(badge?.classList.contains("survives")).toBe(true);
    expect(badge?.textContent?.trim()).toBe("4");
    expect(badge?.querySelector(".survives-icon svg")).not.toBeNull();
    expect(badge?.getAttribute("title")).toContain("but it has indestructible");
    expect(badge?.getAttribute("aria-label")).toBe("damage, indestructible");
  });

  it("is the plain count on a creature that can die", () => {
    const { container } = render(Card as never, { card: solphim({ damage_marked: 2 }) } as never);
    const badge = container.querySelector(".badge.damage");
    expect(badge?.classList.contains("survives")).toBe(false);
    expect(badge?.querySelector(".survives-icon")).toBeNull();
    expect(badge?.getAttribute("title")).toBe("2 damage marked");
  });

  it("is absent with no damage marked", () => {
    const { container } = render(Card as never, { card: solphim({ damage_marked: 0 }) } as never);
    expect(container.querySelector(".badge.damage")).toBeNull();
  });
});

describe("damage in the hover panel", () => {
  it("says the damage was lethal but the creature is indestructible", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set(solphim(indestructible));
    await settle();
    const chip = container.querySelector(".state-damage");
    expect(chip?.textContent?.trim()).toBe("4 damage: lethal, but indestructible");
  });

  it("lists plain damage on a creature without indestructible", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set(solphim({ damage_marked: 1 }));
    await settle();
    expect(container.querySelector(".state-damage")?.textContent?.trim()).toBe("1 damage");
  });
});
