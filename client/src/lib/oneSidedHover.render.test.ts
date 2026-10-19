// @vitest-environment jsdom
//
// #4109 — an adventurer, prepare or split card has ONE image and no
// top-level oracle text on Scryfall: the text is per face. The hover
// zoom lists both halves' rules text, and asks for no second scan
// (there is no ?face=1 to fetch), no inset and no Shift flip.

import { describe, it, expect, beforeEach, afterEach } from "vitest";

import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import { cardImageURL } from "./cardImage";
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
  turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
} as unknown as GameView;

const ADV_ID = "adv-1";
const PREP_ID = "prep-1";

// Scryfall's view: a joined top-level name and NO oracle_text.
const META: Record<string, unknown> = {
  [ADV_ID]: {
    id: ADV_ID,
    name: "Bonecrusher Giant // Stomp",
    card_faces: [{ name: "Bonecrusher Giant" }, { name: "Stomp" }],
  },
};

const adventurer = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "bg",
    name: "Bonecrusher Giant",
    owner: "me",
    controller: "me",
    known_by_you: true,
    scryfall_id: ADV_ID,
    layout: "adventure",
    active_face: 0,
    type_line: "Creature — Giant",
    faces: [
      {
        name: "Bonecrusher Giant",
        type_line: "Creature — Giant",
        mana_cost: "{2}{R}",
        oracle_text:
          "Whenever this creature becomes the target of a spell, it deals 2 damage to that spell's controller.",
        power: 4,
        toughness: 3,
      },
      {
        name: "Stomp",
        type_line: "Instant — Adventure",
        mana_cost: "{1}{R}",
        oracle_text: "Damage can't be prevented this turn. Stomp deals 2 damage to any target.",
      },
    ],
    ...extra,
  }) as unknown as CardView;

const preparer = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "sc",
    name: "Skycoach Conductor",
    owner: "me",
    controller: "me",
    known_by_you: true,
    scryfall_id: PREP_ID,
    layout: "prepare",
    active_face: 0,
    type_line: "Creature — Bird Pilot",
    faces: [
      {
        name: "Skycoach Conductor",
        type_line: "Creature — Bird Pilot",
        mana_cost: "{2}{U}",
        oracle_text: "Flash\nFlying, vigilance",
      },
      {
        name: "All Aboard",
        type_line: "Instant",
        mana_cost: "{U}",
        oracle_text: "Tap target creature.",
      },
    ],
    ...extra,
  }) as unknown as CardView;

let realFetch: unknown;
beforeEach(() => {
  hoveredCard.set(null);
  const g = globalThis as Record<string, unknown>;
  realFetch = g.fetch;
  g.fetch = (url: string) => {
    const id = String(url).split("/").pop() ?? "";
    const meta = META[id] ?? { id, name: "?" };
    return Promise.resolve({ ok: true, json: () => Promise.resolve(meta) } as unknown as Response);
  };
});
afterEach(() => {
  hoveredCard.set(null);
  window.dispatchEvent(new Event("blur"));
  cleanup();
  (globalThis as Record<string, unknown>).fetch = realFetch;
});

function show(card: CardView): HTMLElement {
  const { container } = render(HoverZoomOverlay as never, { view } as never);
  hoveredCard.set(card);
  flushSync();
  return container;
}

const otherSide = (c: HTMLElement) => c.querySelector<HTMLElement>(".other-side");

describe("hover zoom of an adventurer or prepare card (#4109)", () => {
  it("reads the creature's text in the usual place and the Adventure's below it", () => {
    const c = show(adventurer());
    expect(c.querySelector(".info > .oracle")?.textContent).toContain("becomes the target");
    expect(otherSide(c)?.querySelector("h3")?.textContent).toBe("Adventure");
    expect(otherSide(c)?.querySelector(".title")?.textContent).toBe("Stomp");
    expect(otherSide(c)?.querySelector(".other-oracle")?.textContent).toContain(
      "deals 2 damage to any target",
    );
  });

  it("reads a prepare card's creature and its prepare spell", () => {
    const c = show(preparer());
    expect(c.querySelector(".info > .oracle")?.textContent).toContain("Flying, vigilance");
    expect(otherSide(c)?.querySelector("h3")?.textContent).toBe("Prepare spell");
    expect(otherSide(c)?.querySelector(".title")?.textContent).toBe("All Aboard");
    expect(otherSide(c)?.querySelector(".other-oracle")?.textContent).toBe("Tap target creature.");
  });

  it("names the creature as the other half when the Adventure is the face up", () => {
    const c = show(adventurer({ active_face: 1, name: "Stomp" }));
    expect(otherSide(c)?.querySelector("h3")?.textContent).toBe("Creature");
    expect(otherSide(c)?.querySelector(".title")?.textContent).toBe("Bonecrusher Giant");
  });

  it("shows one image: no inset, no flip hint, no Shift flip, no ?face=1 request", () => {
    const c = show(adventurer());
    expect(c.querySelector(".other-face")).toBeNull();
    expect(c.querySelector(".flip-hint")).toBeNull();
    const before = c.querySelector<HTMLImageElement>(".scan > img")?.getAttribute("src");
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Shift", shiftKey: true }));
    flushSync();
    expect(c.querySelector<HTMLImageElement>(".scan > img")?.getAttribute("src")).toBe(before);
    expect(before).not.toContain("face=");
  });

  it("asks for the whole card even when its second face is the one up", () => {
    expect(cardImageURL(adventurer({ active_face: 1 }), "normal")).not.toContain("face=");
    expect(cardImageURL(preparer({ active_face: 1 }), "normal", 1)).not.toContain("face=");
  });

  it("lists both halves of a split card too, still as one sideways image", () => {
    const c = show({
      instance_id: "fi",
      name: "Fire // Ice",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: "split-1",
      layout: "split",
      active_face: 0,
      faces: [
        {
          name: "Fire",
          type_line: "Instant",
          oracle_text: "Fire deals 2 damage divided as you choose",
        },
        { name: "Ice", type_line: "Instant", oracle_text: "Tap target permanent." },
      ],
    } as unknown as CardView);
    expect(c.querySelector(".info > .oracle")?.textContent).toContain("Fire deals 2 damage");
    expect(otherSide(c)?.querySelector("h3")?.textContent).toBe("Other half");
    expect(otherSide(c)?.querySelector(".other-oracle")?.textContent).toBe("Tap target permanent.");
    expect(c.querySelector(".other-face")).toBeNull();
    expect(c.querySelector(".scan > img.sideways")).not.toBeNull();
  });
});
