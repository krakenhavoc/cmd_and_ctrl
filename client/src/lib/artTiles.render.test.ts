// @vitest-environment jsdom
//
// #2209 (was #1954's experimental "card art only"). Battlefield cards
// are art tiles by default: Scryfall's art_crop with a name strip, the
// P/T or loyalty badge, counters, status marks and keyword chips. The
// viewer's own hand keeps the full card unless `display.handArt` is on.
// The hover zoom, the stack, face-down cards, opponents' hands and every
// Card that does not opt in keep the size they always used.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import Hand from "./components/board/Hand.svelte";
import Card from "./components/board/Card.svelte";
import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import { settings, updateSettings, defaultSettings } from "./settings";
import { tableImageSize } from "./cardImage";
import { hoveredCard } from "./cardTypes";
import type { CardView, GameView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const ID = "11111111-1111-1111-1111-111111111111";
const bear = (extra: Partial<CardView> = {}): CardView => ({
  instance_id: "bear",
  name: "Grizzly Bears",
  owner: "me",
  controller: "me",
  type_line: "Creature — Bear",
  power: 2,
  toughness: 2,
  scryfall_id: ID,
  known_by_you: true,
  ...extra,
});

function setArt(battlefield: boolean, hand: boolean) {
  updateSettings("display", "battlefieldArt", battlefield);
  updateSettings("display", "handArt", hand);
  flushSync();
}

let realFetch: unknown;
beforeEach(() => {
  localStorage.clear();
  const g = globalThis as Record<string, unknown>;
  realFetch = g.fetch;
  // The hover panel asks for oracle text; nothing here reads it.
  g.fetch = () =>
    Promise.resolve({ ok: true, json: () => Promise.resolve({}) } as unknown as Response);
});
afterEach(() => {
  const d = defaultSettings().display;
  setArt(d.battlefieldArt, d.handArt);
  hoveredCard.set(null);
  cleanup();
  (globalThis as Record<string, unknown>).fetch = realFetch;
});

const srcs = (c: HTMLElement) =>
  [...c.querySelectorAll<HTMLImageElement>("img:not(.back-img)")].map((i) => i.getAttribute("src"));
const strip = (c: HTMLElement) => c.querySelector<HTMLElement>(".art-name");

function row(cards: CardView[]) {
  return render(
    BattlefieldRow as never,
    {
      label: "creatures",
      cards,
      viewerID: "me",
      attachmentsByHost: {},
      onCardClick: () => {},
    } as never,
  );
}
function hand(cards: CardView[], isSelf = true) {
  return render(
    Hand as never,
    {
      hand: { kind: "hand", owner: "me", count: cards.length, cards },
      isSelf,
      viewerID: "me",
    } as never,
  );
}

describe("art tiles by default (#2209)", () => {
  it("defaults: art on the battlefield, full cards in the hand", () => {
    expect(defaultSettings().display.battlefieldArt).toBe(true);
    expect(defaultSettings().display.handArt).toBe(false);
    expect(get(settings).display.battlefieldArt).toBe(true);
    expect(get(settings).display.handArt).toBe(false);
  });

  it("a battlefield card draws art_crop with a name strip; your hand draws the full card", () => {
    const r = row([bear()]).container;
    expect(srcs(r)).toEqual([`/cards/${ID}/image?size=art_crop`]);
    expect(strip(r)?.textContent).toBe("Grizzly Bears");
    expect(r.querySelector(".card")!.classList.contains("art-tile")).toBe(true);

    const h = hand([bear()]).container;
    expect(srcs(h)).toEqual([`/cards/${ID}/image?size=small`]);
    expect(strip(h)).toBeNull();
  });

  it("the tile keeps the P/T badge, counters, status marks and keyword chips", () => {
    const r = row([
      bear({
        abilities: ["flying", "deathtouch"],
        counters: { "+1/+1": 1 },
        goaded_by: "them",
        power: 3,
        toughness: 3,
      }),
    ]).container;
    expect(r.querySelector(".badge.pt")?.textContent?.trim()).toBe("3/3");
    expect(r.querySelector(".pips")).not.toBeNull();
    expect(r.querySelector(".badge.goad")).not.toBeNull();
    const chips = [...r.querySelectorAll(".keyword-row .kw-badge")].map((b) =>
      b.getAttribute("aria-label"),
    );
    expect(chips).toEqual(["Flying", "Deathtouch"]);
  });

  it("a planeswalker's tile keeps its loyalty badge, and a land gets a strip too", () => {
    const pw = row([
      bear({
        instance_id: "pw",
        name: "Chandra",
        type_line: "Legendary Planeswalker — Chandra",
        counters: { loyalty: 4 },
      }),
    ]).container;
    expect(pw.querySelector(".badge.loyalty")?.textContent?.trim()).toBe("4");
    cleanup();
    const land = row([
      bear({ instance_id: "isl", name: "Island", type_line: "Basic Land — Island" }),
    ]).container;
    expect(strip(land)?.textContent).toBe("Island");
  });

  it("a planeswalker that is also a creature shows loyalty and power/toughness together (#2046)", () => {
    const gideon = row([
      bear({
        instance_id: "gideon",
        name: "Gideon, Ally of Zendikar",
        type_line: "Legendary Creature Planeswalker — Gideon Human Soldier Ally",
        power: 5,
        toughness: 5,
        counters: { loyalty: 5 },
      }),
    ]).container;
    expect(gideon.querySelector(".badge.loyalty")?.textContent?.trim()).toBe("5");
    expect(gideon.querySelector(".badge.loyalty")?.classList.contains("with-pt")).toBe(true);
    expect(gideon.querySelector(".badge.pt")?.textContent?.trim()).toBe("5/5");
  });

  it("keeps role=button and the card's name as its accessible name", () => {
    const el = row([bear()]).container.querySelector(".card") as HTMLElement;
    expect(el.getAttribute("role")).toBe("button");
    expect(el.getAttribute("title")).toBe("Grizzly Bears");
    expect(el.getAttribute("aria-label")).toBe("Grizzly Bears");
    // The strip is decoration: the name is already the card's name.
    expect(strip(el)?.getAttribute("aria-hidden")).toBe("true");
  });

  it("each setting is its own: battlefield off, hand on", () => {
    setArt(false, true);
    const r = row([bear()]).container;
    expect(srcs(r)).toEqual([`/cards/${ID}/image?size=small`]);
    expect(strip(r)).toBeNull();
    cleanup();
    const h = hand([bear()]).container;
    expect(srcs(h)).toEqual([`/cards/${ID}/image?size=art_crop`]);
    expect(strip(h)?.textContent).toBe("Grizzly Bears");
  });

  it("an opponent's hand is untouched", () => {
    setArt(true, true);
    const c = hand([bear({ known_by_you: false })], false).container;
    expect(srcs(c)).toEqual([]);
    expect(strip(c)).toBeNull();
  });

  it("a face-down permanent never shows art or a name strip", () => {
    // Your own morph: the real face at the usual size, never the crop.
    const own = row([bear({ face_down: true, face_visible: true })]).container;
    expect(srcs(own).every((s) => !s?.includes("art_crop"))).toBe(true);
    expect(strip(own)).toBeNull();
    cleanup();
    // Someone else's: the back, nothing else.
    const theirs = row([
      bear({
        face_down: true,
        known_by_you: false,
        name: "",
        scryfall_id: undefined,
        controller: "them",
      }),
    ]).container;
    expect(srcs(theirs)).toEqual([]);
    expect(theirs.querySelector(".back-img")).not.toBeNull();
    expect(strip(theirs)).toBeNull();
  });

  it("a token with no art falls back to its name, as before", () => {
    const r = row([
      bear({ instance_id: "tok", name: "Soldier", is_token: true, scryfall_id: undefined }),
    ]).container;
    expect(srcs(r)).toEqual([]);
    expect(r.querySelector(".name-fallback")?.textContent).toBe("Soldier");
    expect(strip(r)).toBeNull();
  });

  it("the stack and every Card that does not opt in keep the full card", () => {
    const stack = render(Card as never, { card: bear() } as never);
    expect(srcs(stack.container)).toEqual([`/cards/${ID}/image?size=small`]);
    expect(strip(stack.container)).toBeNull();
  });

  it("the hover zoom still requests the full card", () => {
    const view = {
      id: "g",
      state: "active",
      seats: [],
      battlefield: { kind: "battlefield", count: 0, cards: [] },
      stack: { kind: "stack", count: 0, cards: [] },
      exile: { kind: "exile", count: 0, cards: [] },
      turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main", step: "main" },
    } as unknown as GameView;
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set(bear());
    flushSync();
    const imgs = srcs(container);
    expect(imgs).toContain(`/cards/${ID}/image?size=normal`);
    expect(imgs.every((s) => !s?.includes("art_crop"))).toBe(true);
  });
});

describe("tableImageSize", () => {
  it("only swaps a face-up card when the setting is on", () => {
    expect(tableImageSize({}, "small", false)).toBe("small");
    expect(tableImageSize({}, "small", true)).toBe("art_crop");
    expect(tableImageSize({ face_down: true }, "small", true)).toBe("small");
  });

  it("a card with no Scryfall ID has no URL to fall back from (the server serves normal for a missing crop)", async () => {
    const { cardImageURL } = await import("./cardImage");
    expect(cardImageURL({ scryfall_id: undefined } as never, "art_crop")).toBeNull();
  });
});
