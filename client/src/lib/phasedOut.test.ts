// phasedOut.test.ts — #1199, ADR 0084, the pure half.
//
// The decision "which cards belong on whose row" is worth testing
// without mounting anything, for cardBack.ts's reason. What is
// load-bearing here is CONTROLLER, not owner: CR 702.26d says the
// phasing event does not change control, so a creature stolen and
// then phased out is still on the thief's side of the table and that
// is where the player expects to watch for it coming back.

import { describe, it, expect } from "vitest";

import type { CardView, GameView, ZoneView } from "./protocol";
import { isPhasedOut, phasedOutCards, phasedOutByController } from "./phasedOut";

function card(id: string, owner: string, controller: string): CardView {
  return {
    instance_id: id,
    name: id,
    owner,
    controller,
    phased_out: true,
  } as unknown as CardView;
}

function view(cards: CardView[]): GameView {
  const zone: ZoneView = {
    kind: "phased_out",
    count: cards.length,
    cards,
  } as unknown as ZoneView;
  return { phased_out: zone } as unknown as GameView;
}

describe("isPhasedOut", () => {
  it("is true only for an explicit true", () => {
    expect(isPhasedOut(card("a", "me", "me"))).toBe(true);
  });

  it("is false for the ABSENT field, which is how the wire spells false", () => {
    const plain = { instance_id: "b", name: "b", owner: "me", controller: "me" };
    expect(isPhasedOut(plain as unknown as CardView)).toBe(false);
  });
});

describe("phasedOutCards", () => {
  it("returns the zone's cards", () => {
    expect(phasedOutCards(view([card("a", "me", "me")]))).toHaveLength(1);
  });

  it("tolerates a server that has no such zone", () => {
    expect(phasedOutCards({} as unknown as GameView)).toEqual([]);
  });
});

describe("phasedOutByController", () => {
  it("groups by CONTROLLER, not owner — CR 702.26d leaves control alone", () => {
    const stolen = card("stolen", "victim", "thief");
    const mine = card("mine", "thief", "thief");
    const theirs = card("theirs", "victim", "victim");

    const byController = phasedOutByController(view([stolen, mine, theirs]));

    expect(byController.get("thief")?.map((c) => c.instance_id)).toEqual(["stolen", "mine"]);
    expect(byController.get("victim")?.map((c) => c.instance_id)).toEqual(["theirs"]);
    // The owner of the stolen creature does not get it back on their
    // row just because they own it.
    expect(byController.get("victim")).toHaveLength(1);
  });

  it("is empty for a board with nothing phased out", () => {
    expect(phasedOutByController(view([])).size).toBe(0);
  });
});
