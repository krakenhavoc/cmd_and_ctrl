// @vitest-environment jsdom
//
// #3518 — a split card (Fire // Ice, aftermath, a Room) has one image
// with its text sideways. The hover zoom shows it whole, turned a
// quarter-turn, with no second-face inset (both halves are that image).

import { describe, it, expect, beforeEach, afterEach } from "vitest";

import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import { hoveredCard } from "./cardTypes";
import type { CardView, GameView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const view = {
  id: "g",
  state: "active",
  seats: [],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main", step: "main" },
} as unknown as GameView;

const card = (layout: string, id: string): CardView =>
  ({
    instance_id: "c1",
    name: "Unholy Annex",
    owner: "me",
    controller: "me",
    known_by_you: true,
    scryfall_id: id,
    layout,
    active_face: 0,
    type_line: "Enchantment — Room",
    faces: [
      { name: "Unholy Annex", type_line: "Enchantment — Room" },
      { name: "Ritual Chamber", type_line: "Enchantment — Room" },
    ],
  }) as unknown as CardView;

let realFetch: unknown;
beforeEach(() => {
  hoveredCard.set(null);
  const g = globalThis as Record<string, unknown>;
  realFetch = g.fetch;
  g.fetch = () => new Promise(() => {});
});
afterEach(() => {
  cleanup();
  (globalThis as Record<string, unknown>).fetch = realFetch;
});

describe("hover zoom of a split card", () => {
  it("turns the whole image and shows no second-face inset", () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set(card("split", "room-1"));
    flushSync();
    expect(container.querySelector(".scan.sideways img.sideways")).not.toBeNull();
    expect(container.querySelector(".other-face")).toBeNull();
  });

  it("leaves a modal DFC upright with its inset", () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set(card("modal_dfc", "mdfc-1"));
    flushSync();
    expect(container.querySelector(".sideways")).toBeNull();
    expect(container.querySelector(".other-face")).not.toBeNull();
  });
});
