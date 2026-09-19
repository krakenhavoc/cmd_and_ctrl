import { describe, it, expect } from "vitest";

import { cardAsFace, needsFacePicker } from "./faces";
import { applyCastChoices, castLocksXAtZero, hasXCost } from "./targeting";
import type { CardView } from "./protocol";

// impulseCast.test.ts — #874, the other half. The render test pins
// that the exile button reaches the Board's cast chain; this pins what
// the chain then sees, because routing a cast through a chain that
// cannot read the card would buy nothing.
//
// Each case is one `after*` seam's own question, asked of the card the
// modal hands up:
//
//   afterCastCosts  hasXCost && !castLocksXAtZero   -> the X picker
//   handlePlayCard  needsFacePicker                 -> the face picker
//   continueCast    card.target_mode                -> targeting
//
// The bug was never that these read the wrong thing; it was that
// nothing asked them at all.

function exiled(extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "impulsed",
    name: "Fireball",
    owner: "them",
    controller: "them",
    type_line: "Sorcery",
    mana_cost: "{X}{R}",
    target_mode: "any",
    ...extras,
  };
}

describe("an impulse cast for the PRINTED cost", () => {
  // Light Up the Stage, Etali, Prosper: "you may cast it", no price
  // named, so the {X} is genuinely being paid and X > 0 is legal. The
  // server says so by leaving x_locked_at_zero off the grant.
  const printedCost = exiled({ exile_play: { player: "me" } });

  it("opens the X picker", () => {
    expect(hasXCost(printedCost)).toBe(true);
    expect(castLocksXAtZero(printedCost, undefined)).toBe(false);
  });

  it("opens the target picker", () => {
    expect(printedCost.target_mode).toBe("any");
  });

  it("sends the zone and the announced X together", () => {
    const params: Record<string, unknown> = { instance_id: printedCost.instance_id };
    applyCastChoices(params, { fromZone: "exile", xValue: 4 });
    expect(params).toEqual({ instance_id: "impulsed", from_zone: "exile", x_value: 4 });
  });
});

describe("a FREE impulse cast", () => {
  // A cascade hit or a Siege's transformed cast is granted at {0}, so
  // CR 107.3b leaves 0 as the only legal X and the picker must not
  // open. The server computes that with the predicate it will judge
  // the cast by and ships the answer on the grant (#831).
  const free = exiled({
    exile_play: { player: "me", cast_only: true, cost_override: "{0}", x_locked_at_zero: true },
  });

  it("skips the X picker", () => {
    expect(hasXCost(free)).toBe(true);
    expect(castLocksXAtZero(free, undefined)).toBe(true);
  });
});

describe("an impulse cast of a granted FACE", () => {
  // A defeated Siege is exiled battle-side-up under a grant for face
  // 1. The Board applies cardAsFace before the rest of the chain, so
  // every prompt reads the half being cast — and the grant has to
  // survive that swap, or the free cast would suddenly offer an X
  // picker for the front face's {X}.
  const siege = exiled({
    name: "Invasion of New Phyrexia",
    type_line: "Battle — Siege",
    layout: "transform",
    exile_play: { player: "me", cast_only: true, cost_override: "{0}", x_locked_at_zero: true },
    faces: [
      { name: "Invasion of New Phyrexia", type_line: "Battle — Siege", mana_cost: "{X}{W}{U}" },
      { name: "Teferi Akosa of Zhalfir", type_line: "Legendary Planeswalker — Teferi" },
    ],
  });

  it("is never sent to the face picker — the grant leaves no choice", () => {
    // `transform`, not `modal_dfc`: the picker is for a card whose
    // faces are independently playable (CR 712.12a), and this one's
    // back face is reachable only through the grant.
    expect(needsFacePicker(siege)).toBe(false);
  });

  it("reads the granted face, and keeps the grant across the swap", () => {
    const back = cardAsFace(siege, 1);
    expect(back.name).toBe("Teferi Akosa of Zhalfir");
    expect(hasXCost(back)).toBe(false);
    expect(castLocksXAtZero(back, undefined)).toBe(true);
  });

  it("sends the face alongside the zone", () => {
    const params: Record<string, unknown> = { instance_id: siege.instance_id };
    applyCastChoices(params, { fromZone: "exile", face: 1 });
    expect(params).toEqual({ instance_id: "impulsed", from_zone: "exile", face: 1 });
  });
});
