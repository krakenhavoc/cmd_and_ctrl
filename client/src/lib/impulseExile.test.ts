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
