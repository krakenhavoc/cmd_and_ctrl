import { describe, it, expect } from "vitest";

import { castableFromZone } from "./zoneBrowser.logic";
import { applyCastChoices } from "./targeting";
import type { CardView } from "./protocol";

// castFromZone.test.ts — S29. Two small surfaces, both of which turn
// into a wrong button or a wrong payload if they drift:
//
//   - castableFromZone decides whether the zone browser offers a
//     cast at all. `castable_here` is PUBLIC (the graveyard is a
//     public zone), so the ownership check here is load-bearing
//     rather than a duplicate of the server's — without it the
//     browser paints a button on an opponent's card that the server
//     answers with "card not found".
//   - applyCastChoices is the single place that knows the wire
//     names, and `from_zone` is the one that decides which pile the
//     server reaches into.

function inYard(extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "looting",
    name: "Faithless Looting",
    owner: "me",
    controller: "me",
    type_line: "Sorcery",
    ...extras,
  };
}

describe("castableFromZone — who gets the graveyard cast button", () => {
  it("offers the cast to the graveyard's owner when the server marked the card", () => {
    expect(castableFromZone(inYard({ castable_here: true }), "graveyard", "me", "me")).toBe(true);
  });

  it("withholds it from everyone else, even though the bit is public", () => {
    const card = inYard({ castable_here: true });
    expect(castableFromZone(card, "graveyard", "them", "me")).toBe(false);
    expect(castableFromZone(card, "graveyard", null, "me")).toBe(false);
  });

  // #1022 / #1037. A permission names an OBJECT, so a card in an
  // opponent's graveyard is castable by the seat that holds one —
  // Wrexial's "cast target instant or sorcery card from that player's
  // graveyard". The server computes that holder's own offers, targets
  // and gate for their frame alone and names them in the public
  // `exile_play`; the ownership gate above would have dropped the
  // button on the floor.
  it("offers it to a seat the grant names, in somebody else's graveyard", () => {
    const granted = inYard({ castable_here: true, exile_play: { player: "them" } });
    expect(castableFromZone(granted, "graveyard", "them", "me")).toBe(true);
    // The pile's owner keeps their own answer, and a third seat gets
    // neither: the stamps on their frame are the owner's public ones.
    expect(castableFromZone(granted, "graveyard", "me", "me")).toBe(true);
    expect(castableFromZone(granted, "graveyard", "third", "me")).toBe(false);
  });

  it("withholds it from a card the server did not mark", () => {
    expect(castableFromZone(inYard(), "graveyard", "me", "me")).toBe(false);
    expect(castableFromZone(inYard({ castable_here: false }), "graveyard", "me", "me")).toBe(false);
  });

  it("is graveyard-only — exile keeps its own grant-keyed button", () => {
    const card = inYard({ castable_here: true });
    expect(castableFromZone(card, "exile", "me", "me")).toBe(false);
    expect(castableFromZone(card, "command", "me", "me")).toBe(false);
    expect(castableFromZone(card, "stack", "me", "me")).toBe(false);
  });
});

describe("applyCastChoices — from_zone on the wire", () => {
  it("omits from_zone for a hand cast", () => {
    const params: Record<string, unknown> = { instance_id: "x" };
    applyCastChoices(params, {});
    expect(params).toEqual({ instance_id: "x" });
  });

  it("sends the zone alongside the rest of the announce-time choices", () => {
    const params: Record<string, unknown> = { instance_id: "x" };
    applyCastChoices(params, { fromZone: "graveyard", altCost: "flashback" });
    expect(params).toEqual({
      instance_id: "x",
      from_zone: "graveyard",
      alternative_cost: "flashback",
    });
  });
});
