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

  // #1055. "Whose answer is the bit" used to be a question this
  // function had to answer, because `castable_here` was PUBLIC and
  // carried the LIBRARY OWNER's answer: two seats at a table with
  // Oracles of Mul Daya both saw the other's revealed top card marked
  // playable. It takes the viewer's and the owner's seat no longer —
  // the server stamps the bit for the seat it is projecting, so a
  // frame that carries it is a frame that may play the card.
  //
  // The server holds that half now, in
  // TestTwoHoldersEachSeeTheirOwnGraveyardCastStamps and
  // TestTheLibraryOwnerAndBystandersGetNoForeignLibraryStamps
  // (server/internal/protocol). What is left here is that the reader
  // reads the bit and nothing beside it.
  describe("whose answer the bit is", () => {
    it("plays a top card the server marked for THIS frame", () => {
      const top = card({ name: "Their Top", known_by_you: true, castable_here: true });
      expect(libraryTopPlayable(library([top]))).toBe(true);
    });

    it("does not play one it did not mark, grant or no grant", () => {
      // `exile_play` is public and names the seat the permission was
      // granted to. It is not the answer to "may I play this" any
      // more, so a bystander's frame — the card, the public grant, no
      // bit — is not a button.
      const bystander = card({
        name: "Their Top",
        known_by_you: true,
        exile_play: { player: "them" },
      });
      expect(libraryTopPlayable(library([bystander]))).toBe(false);
    });
  });
});
