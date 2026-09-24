import { describe, expect, it } from "vitest";

import {
  libraryTopActionLabel,
  libraryTopPlayable,
  libraryTopSpecialActions,
  visibleLibraryTop,
} from "./libraryTop";
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

// #1440: the verb for the PileBar affordance. A land is played, not
// cast (CR 305.1, CR 116.2a) — the same wording nit the graveyard's
// flashback button has to mind (ZoneBrowserModal's castLabelFor).
describe("libraryTopActionLabel", () => {
  it("is null when the top isn't visible or isn't playable", () => {
    expect(libraryTopActionLabel(undefined)).toBeNull();
    expect(libraryTopActionLabel(library([], 40))).toBeNull();
    // Oracle of Mul Daya reveals a sorcery to the whole table but
    // opens only lands: visible, not playable, no button.
    const revealedOnly = card({
      name: "Top Sorcery",
      type_line: "Sorcery",
      known_by_you: true,
    });
    expect(libraryTopActionLabel(library([revealedOnly]))).toBeNull();
  });

  it("says play for a land", () => {
    const land = card({
      name: "Top Forest",
      type_line: "Basic Land — Forest",
      known_by_you: true,
      castable_here: true,
    });
    expect(libraryTopActionLabel(library([land]))).toBe("play");
  });

  it("says cast for anything else", () => {
    const spell = card({
      name: "Top Sorcery",
      type_line: "Sorcery",
      known_by_you: true,
      castable_here: true,
    });
    expect(libraryTopActionLabel(library([spell]))).toBe("cast");
  });
});

// #1391: Fblthp, Lost on the Range's plot from the top of the library.
// The server stamps the rows; this only picks the available ones and
// builds the `special_action` params.
describe("libraryTopSpecialActions", () => {
  it("is empty when the top isn't visible or offers nothing available", () => {
    expect(libraryTopSpecialActions(undefined)).toEqual([]);
    expect(libraryTopSpecialActions(library([], 40))).toEqual([]);
    const hidden = card({
      special_actions: [{ kind: "plot", label: "Plot {1}{G}", cost: "{1}{G}", available: true }],
    });
    expect(libraryTopSpecialActions(library([hidden]))).toEqual([]);
    const notNow = card({
      name: "Grizzly Bears",
      known_by_you: true,
      special_actions: [{ kind: "plot", label: "Plot {1}{G}", cost: "{1}{G}", available: false }],
    });
    expect(libraryTopSpecialActions(library([notNow]))).toEqual([]);
  });

  it("turns one available row into a 'plot' pill with the special_action params", () => {
    const bears = card({
      instance_id: "bears",
      name: "Grizzly Bears",
      known_by_you: true,
      special_actions: [{ kind: "plot", label: "Plot {1}{G}", cost: "{1}{G}", available: true }],
    });
    expect(libraryTopSpecialActions(library([bears]))).toEqual([
      {
        key: "plot:{1}{G}",
        text: "plot",
        label: "Plot {1}{G}",
        params: { card_id: "bears", kind: "plot", strict: true, auto_tap: true, cost: "{1}{G}" },
      },
    ]);
  });

  it("names the price on each pill when the card offers the kind twice", () => {
    const djinn = card({
      instance_id: "djinn",
      name: "Djinn of Fool's Fall",
      known_by_you: true,
      special_actions: [
        { kind: "plot", label: "Plot {3}{U}", cost: "{3}{U}", available: true },
        {
          kind: "plot",
          label: "Plot {4}{U}",
          cost: "{4}{U}",
          charged_cost: "{2}{U}",
          available: true,
        },
      ],
    });
    const pills = libraryTopSpecialActions(library([djinn]));
    expect(pills.map((p) => p.text)).toEqual(["plot {3}{U}", "plot {2}{U}"]);
    // The price shown is the charged one; the price SENT is the
    // printed one, which is what the server matches on.
    expect(pills.map((p) => p.params.cost)).toEqual(["{3}{U}", "{4}{U}"]);
  });
});
