// @vitest-environment jsdom
//
// ADR 0069 decision 6, the client half: what a face-down object looks
// like to each side of the table.
//
// `showsCardBack` is unit-tested in cardBack.test.ts and decides which
// of Card.svelte's two arms runs. What only exists once the markup is
// rendered — and so is what this file is for — is the BADGE: the table
// has to be able to tell a morph from an ordinary card back, and the
// controller of their own face-down permanent has to be able to tell
// that the card they are looking at is face down to everyone else.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function mount(card: CardView) {
  return render(Card as never, { card } as Record<string, unknown>);
}

// What the wire hands a seat that may NOT look: the public CR 708.2
// body, the kind, and nothing that names the card.
const opponentsMorph = (): CardView =>
  ({
    instance_id: "morph",
    name: "",
    owner: "them",
    controller: "them",
    type_line: "Creature",
    power: 2,
    toughness: 2,
    face_down: true,
    face_down_kind: "morphed",
  }) as unknown as CardView;

// What the wire hands the controller: the same object, plus the art
// and the permission to look at it (CR 708.5).
const myManifest = (): CardView =>
  ({
    instance_id: "mine",
    name: "",
    owner: "me",
    controller: "me",
    scryfall_id: "aaaa-bbbb",
    type_line: "Creature",
    power: 2,
    toughness: 2,
    face_down: true,
    face_down_kind: "manifested",
    face_visible: true,
    known_by_you: true,
  }) as unknown as CardView;

describe("a face-down card the viewer may not look at", () => {
  it("draws a card back", () => {
    const { container } = mount(opponentsMorph());
    expect(container.querySelector("img.back-img")).not.toBeNull();
    expect(container.querySelector(".card")?.classList.contains("face-down")).toBe(true);
  });

  it("labels the back with the public kind", () => {
    const { container } = mount(opponentsMorph());
    expect(container.querySelector(".badge.face-down")?.textContent?.trim()).toBe("MORPHED");
  });

  it("never requests the card's art", () => {
    const { container } = mount(opponentsMorph());
    const srcs = [...container.querySelectorAll("img")].map((i) => i.getAttribute("src"));
    expect(srcs.every((s) => s?.startsWith("/card-back"))).toBe(true);
  });
});

describe("a face-down card the viewer may look at", () => {
  it("draws the real face, not a back (CR 708.5)", () => {
    const { container } = mount(myManifest());
    expect(container.querySelector("img.back-img")).toBeNull();
  });

  it("badges it as face down, so the controller knows the table sees a back", () => {
    const { container } = mount(myManifest());
    expect(container.querySelector(".badge.face-down")?.textContent?.trim()).toBe("MANIFESTED");
  });
});

describe("an ordinary card", () => {
  it("carries no face-down badge", () => {
    const { container } = mount({
      instance_id: "bolt",
      name: "Lightning Bolt",
      owner: "me",
      controller: "me",
      known_by_you: true,
    } as unknown as CardView);
    expect(container.querySelector(".badge.face-down")).toBeNull();
  });
});
