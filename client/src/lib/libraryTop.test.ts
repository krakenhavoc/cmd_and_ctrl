import { describe, expect, it } from "vitest";

import { libraryTopPlayable, visibleLibraryTop } from "./libraryTop";
import type { CardView, ZoneView } from "./protocol";

function card(over: Partial<CardView> = {}): CardView {
  return {
    instance_id: over.instance_id ?? "c1",
    name: "",
    owner: "me",
    controller: "me",
    ...over,
  } as CardView;
}

function library(cards: CardView[], count = cards.length): ZoneView {
  return { kind: "library", owner: "me", count, cards };
}

describe("visibleLibraryTop", () => {
  it("is null for an empty or missing library", () => {
    expect(visibleLibraryTop(undefined)).toBeNull();
    expect(visibleLibraryTop(library([], 40))).toBeNull();
  });

  it("is null when the top card came back redacted", () => {
    expect(visibleLibraryTop(library([card({ known_by_you: false })]))).toBeNull();
    // known but nameless is the shape the redactor leaves behind, and
    // it must not read as visible.
    expect(visibleLibraryTop(library([card({ known_by_you: true })]))).toBeNull();
  });

  it("is the LAST card, which is the top", () => {
    const buried = card({ instance_id: "buried", name: "Buried", known_by_you: true });
    const top = card({ instance_id: "top", name: "Top Card", known_by_you: true });
    expect(visibleLibraryTop(library([buried, top]))?.instance_id).toBe("top");
  });
});

describe("libraryTopPlayable", () => {
  it("needs castable_here as well as visibility", () => {
    // Oracle of Mul Daya reveals the top card to everyone but opens
    // only lands, so a revealed sorcery is visible and unplayable.
    const revealedOnly = card({ name: "Top Sorcery", known_by_you: true });
    expect(visibleLibraryTop(library([revealedOnly]))).not.toBeNull();
    expect(libraryTopPlayable(library([revealedOnly]))).toBe(false);

    const playable = card({ name: "Top Forest", known_by_you: true, castable_here: true });
    expect(libraryTopPlayable(library([playable]))).toBe(true);
  });

  it("is false for a library nobody may look at", () => {
    expect(libraryTopPlayable(library([card({ castable_here: true })]))).toBe(false);
  });
});
