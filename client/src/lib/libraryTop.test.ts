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

  // #1035. `castable_here` on a library top is the LIBRARY OWNER's
  // answer and it is public, so a viewer looking at somebody else's
  // pile has to be told whose it is. The server says so the way it
  // does for a foreign graveyard cast: the public `exile_play` names
  // the holder, and that holder's frame is the only one carrying their
  // stamps.
  describe("whose answer the bit is", () => {
    const top = card({ name: "Their Top", known_by_you: true, castable_here: true });

    it("offers the viewer's own library top", () => {
      expect(libraryTopPlayable(library([top]), "me", "me")).toBe(true);
    });

    it("withholds an opponent's, even though the bit is public", () => {
      expect(libraryTopPlayable(library([top]), "them", "me")).toBe(false);
      expect(libraryTopPlayable(library([top]), null, "me")).toBe(false);
    });

    it("offers it to the seat a cross-seat grant names", () => {
      const granted = card({
        name: "Their Top",
        known_by_you: true,
        castable_here: true,
        exile_play: { player: "them" },
      });
      expect(libraryTopPlayable(library([granted]), "them", "me")).toBe(true);
      // And to nobody else: a third seat that can see a revealed top
      // card gets no button off somebody else's grant.
      expect(libraryTopPlayable(library([granted]), "third", "me")).toBe(false);
    });
  });
});
