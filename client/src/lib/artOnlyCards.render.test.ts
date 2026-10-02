// @vitest-environment jsdom
//
// #1954 — the experimental "card art only" display. A battlefield card
// and the viewer's own hand ask for Scryfall's art_crop; the hover
// zoom, the stack, face-down cards, opponents' hands and every Card
// that does not opt in keep the size they always used. With the
// setting off nothing changes.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import Hand from "./components/board/Hand.svelte";
import Card from "./components/board/Card.svelte";
import { settings, updateSettings, defaultSettings } from "./settings";
import { tableImageSize } from "./cardImage";
import type { CardView } from "./protocol";
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

function setArtOnly(on: boolean) {
  updateSettings("display", "artOnlyCards", on);
  flushSync();
}

beforeEach(() => localStorage.clear());
afterEach(() => {
  setArtOnly(false);
  cleanup();
});

const srcs = (c: HTMLElement) =>
  [...c.querySelectorAll<HTMLImageElement>("img:not(.back-img)")].map((i) => i.getAttribute("src"));

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

describe("card art only (#1954)", () => {
  it("is off by default", () => {
    expect(defaultSettings().display.artOnlyCards).toBe(false);
    expect(get(settings).display.artOnlyCards).toBe(false);
  });

  it("changes nothing while off", () => {
    expect(srcs(row([bear()]).container)).toEqual([`/cards/${ID}/image?size=small`]);
    expect(srcs(hand([bear()]).container)).toEqual([`/cards/${ID}/image?size=small`]);
  });

  it("a battlefield card and your own hand card request art_crop when on", () => {
    setArtOnly(true);
    expect(srcs(row([bear()]).container)).toEqual([`/cards/${ID}/image?size=art_crop`]);
    expect(srcs(hand([bear()]).container)).toEqual([`/cards/${ID}/image?size=art_crop`]);
  });

  it("keeps the card's name as title and aria-label", () => {
    setArtOnly(true);
    const el = row([bear()]).container.querySelector(".card") as HTMLElement;
    expect(el.getAttribute("title")).toBe("Grizzly Bears");
    expect(el.getAttribute("aria-label")).toContain("Grizzly Bears");
  });

  it("an opponent's hand is untouched", () => {
    setArtOnly(true);
    const c = hand([bear({ known_by_you: false })], false).container;
    expect(srcs(c)).toEqual([]);
  });

  it("a face-down card, a stack card and the hover zoom keep their size", () => {
    setArtOnly(true);
    // Face-down permanent the viewer may look at: real art, usual size.
    expect(
      srcs(row([bear({ face_down: true, face_visible: true })]).container).every(
        (s) => !s?.includes("art_crop"),
      ),
    ).toBe(true);
    // A Card that does not opt in (stack, prompts, zoom) never asks for a crop.
    const zoom = render(Card as never, { card: bear(), size: "normal" } as never);
    expect(srcs(zoom.container)).toEqual([`/cards/${ID}/image?size=normal`]);
    const stack = render(Card as never, { card: bear() } as never);
    expect(srcs(stack.container)).toEqual([`/cards/${ID}/image?size=small`]);
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
