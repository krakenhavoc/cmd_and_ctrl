import { describe, it, expect } from "vitest";

import { impulseActionLabel, impulseGrantFor } from "./zoneBrowser.logic";
import type { CardView } from "./protocol";

// impulseExile.test.ts — S21 sub-PR 6. The grant names a player who
// is usually NOT the card's owner, so every case here is about who
// gets the button.

function exiled(extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "loot",
    name: "Stolen Bolt",
    owner: "victim",
    controller: "victim",
    type_line: "Instant",
    ...extras,
  };
}

const mine = { player: "thief" };

describe("impulse exile — who gets the button", () => {
  it("offers the action to the player the grant names, not the owner", () => {
    const card = exiled({ exile_play: mine });
    expect(impulseGrantFor(card, "exile", "thief")).toEqual(mine);
    expect(impulseGrantFor(card, "exile", "victim")).toBeNull();
  });

  it("ignores cards with no grant, and zones that aren't exile", () => {
    expect(impulseGrantFor(exiled(), "exile", "thief")).toBeNull();
    expect(impulseGrantFor(exiled({ exile_play: mine }), "graveyard", "thief")).toBeNull();
    expect(impulseGrantFor(exiled({ exile_play: mine }), "exile", null)).toBeNull();
  });
});

describe("impulse exile — cast vs play", () => {
  it("labels a nonland card 'cast' under either grant", () => {
    expect(impulseActionLabel(exiled({ exile_play: mine }), "exile", "thief")).toBe("cast");
    expect(
      impulseActionLabel(
        exiled({ exile_play: { player: "thief", cast_only: true } }),
        "exile",
        "thief",
      ),
    ).toBe("cast");
  });

  it("strands a land under a cast-only grant, and plays it otherwise", () => {
    const land = { type_line: "Basic Land — Island" };
    expect(
      impulseActionLabel(
        exiled({ ...land, exile_play: { player: "thief", cast_only: true } }),
        "exile",
        "thief",
      ),
    ).toBeNull();
    expect(impulseActionLabel(exiled({ ...land, exile_play: mine }), "exile", "thief")).toBe(
      "play",
    );
  });

  it("offers nothing when there is no grant", () => {
    expect(impulseActionLabel(exiled(), "exile", "thief")).toBeNull();
  });
});

// --- S29 warp: the window has a floor -----------------------------
//
// A warped creature's grant is stamped the moment the end step exiles
// it — on the turn it was warped — and the card says "you may cast it
// from exile ON A LATER TURN". An unbounded grant with no floor would
// light the button up immediately, during the very end step that took
// the creature away.

describe("impulse exile — warp's not-before-turn floor", () => {
  const warped = { player: "thief", not_before_turn: 5 };

  it("withholds the button on the turn the grant was made", () => {
    expect(impulseGrantFor(exiled({ exile_play: warped }), "exile", "thief", 4)).toBeNull();
    expect(impulseActionLabel(exiled({ exile_play: warped }), "exile", "thief", 4)).toBeNull();
  });

  it("offers it from the named turn onwards", () => {
    expect(impulseGrantFor(exiled({ exile_play: warped }), "exile", "thief", 5)).toEqual(warped);
    expect(impulseGrantFor(exiled({ exile_play: warped }), "exile", "thief", 9)).toEqual(warped);
    expect(impulseActionLabel(exiled({ exile_play: warped }), "exile", "thief", 5)).toBe("cast");
  });

  it("leaves floorless grants — impulse exile, airbend — alone", () => {
    expect(impulseGrantFor(exiled({ exile_play: mine }), "exile", "thief", 1)).toEqual(mine);
  });
});
