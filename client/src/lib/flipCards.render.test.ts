// @vitest-environment jsdom
//
// ADR 0151 (#1892): flip cards. A flip card has ONE printed image; its
// flipped half is that image turned upside down. The client always asks
// for face 0 of a flip card, turns a flipped one's image 180° on the
// board (the tile's tap rotation composes with it, so a tapped flipped
// permanent's art sits at 270°), and the hover panel shows the other
// half as the same image turned the other way.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import Card from "./components/board/Card.svelte";
import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import { cardImageURL } from "./cardImage";
import { isFlippedFlipCard, LAYOUT_FLIP, needsFacePicker } from "./faces";
import { hoveredCard } from "./cardTypes";
import type { CardView, GameView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const ID = "22222222-2222-2222-2222-222222222222";

const akki = (extra: Partial<CardView> = {}): CardView => ({
  instance_id: "akki",
  name: "Akki Lavarunner",
  owner: "me",
  controller: "me",
  type_line: "Creature — Goblin Warrior",
  power: 1,
  toughness: 1,
  scryfall_id: ID,
  known_by_you: true,
  layout: LAYOUT_FLIP,
  active_face: 0,
  faces: [
    { name: "Akki Lavarunner", type_line: "Creature — Goblin Warrior", mana_cost: "{3}{R}" },
    { name: "Tok-Tok, Volcano Born", type_line: "Legendary Creature — Goblin Shaman" },
  ] as CardView["faces"],
  ...extra,
});

const tokTok = (extra: Partial<CardView> = {}): CardView =>
  akki({
    name: "Tok-Tok, Volcano Born",
    type_line: "Legendary Creature — Goblin Shaman",
    power: 2,
    toughness: 2,
    active_face: 1,
    ...extra,
  });

let realFetch: unknown;
beforeEach(() => {
  localStorage.clear();
  const g = globalThis as Record<string, unknown>;
  realFetch = g.fetch;
  g.fetch = () =>
    Promise.resolve({ ok: true, json: () => Promise.resolve({}) } as unknown as Response);
});
afterEach(() => {
  hoveredCard.set(null);
  cleanup();
  (globalThis as Record<string, unknown>).fetch = realFetch;
});

describe("cardImageURL for a flip card", () => {
  it("asks for face 0 whatever half is up or asked for", () => {
    const base = `/cards/${ID}/image?size=normal`;
    expect(cardImageURL(akki())).toBe(base);
    expect(cardImageURL(tokTok())).toBe(base);
    expect(cardImageURL(tokTok(), "normal", 1)).toBe(base);
    expect(cardImageURL(akki(), "normal", 1)).toBe(base);
  });

  it("leaves a transformed double-faced card on its active face", () => {
    const dfc = { scryfall_id: ID, layout: "transform", active_face: 1 };
    expect(cardImageURL(dfc)).toBe(`/cards/${ID}/image?size=normal&face=1`);
  });
});

describe("isFlippedFlipCard", () => {
  it("is a flip card on face 1 and nothing else", () => {
    expect(isFlippedFlipCard(tokTok())).toBe(true);
    expect(isFlippedFlipCard(akki())).toBe(false);
    expect(isFlippedFlipCard({ layout: "transform", active_face: 1 })).toBe(false);
    expect(isFlippedFlipCard(null)).toBe(false);
  });

  it("never needs the face picker (CR 710.2)", () => {
    expect(needsFacePicker(akki())).toBe(false);
  });
});

const art = (c: HTMLElement) => c.querySelector<HTMLImageElement>("img:not(.back-img)");

describe("Card.svelte", () => {
  it("turns a flipped flip card's image and nothing else", () => {
    const flipped = render(Card as never, { card: tokTok() } as never);
    expect(art(flipped.container)?.classList.contains("flipped")).toBe(true);
    cleanup();
    const top = render(Card as never, { card: akki() } as never);
    expect(art(top.container)?.classList.contains("flipped")).toBe(false);
  });

  it("keeps the image turned when the permanent is tapped, under the tile's own turn", () => {
    const { container } = render(Card as never, { card: tokTok({ tapped: true }) } as never);
    const tile = container.querySelector<HTMLElement>(".card");
    expect(tile?.classList.contains("tapped")).toBe(true);
    expect(art(container)?.classList.contains("flipped")).toBe(true);
  });

  it("does not turn a transformed double-faced card", () => {
    const { container } = render(
      Card as never,
      { card: akki({ layout: "transform", active_face: 1 }) } as never,
    );
    expect(art(container)?.classList.contains("flipped")).toBe(false);
  });
});

describe("HoverZoomOverlay", () => {
  const view = {
    id: "g",
    state: "active",
    seats: [],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
  } as unknown as GameView;

  function inset(container: HTMLElement) {
    return container.querySelector<HTMLElement>(".other-face:not(.meld-inset)");
  }

  it("shows a flipped permanent turned, with its top half upright as the inset", () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set(tokTok());
    flushSync();
    const main = container.querySelector<HTMLImageElement>(".scan > img");
    expect(main?.classList.contains("flipped")).toBe(true);
    const box = inset(container);
    expect(box?.getAttribute("title")).toBe("Akki Lavarunner");
    const img = box?.querySelector("img");
    expect(img?.getAttribute("src")).toBe(`/cards/${ID}/image?size=normal`);
    expect(img?.classList.contains("flipped")).toBe(false);
  });

  it("shows an unflipped flip card upright, with its flipped half turned as the inset", () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set(akki());
    flushSync();
    const main = container.querySelector<HTMLImageElement>(".scan > img");
    expect(main?.classList.contains("flipped")).toBe(false);
    const box = inset(container);
    expect(box?.getAttribute("title")).toBe("Tok-Tok, Volcano Born");
    const img = box?.querySelector("img");
    expect(img?.getAttribute("src")).toBe(`/cards/${ID}/image?size=normal`);
    expect(img?.classList.contains("flipped")).toBe(true);
  });
});
