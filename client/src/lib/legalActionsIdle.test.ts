// legalActionsIdle.test.ts — #1918. A cast the server marks idle (an
// overloaded Counterflux with no spell to counter) is still legal, so
// the card is still ready; it only draws a muted ring with the server's
// hint as its tooltip, and only when EVERY cast it has is idle.

import { describe, it, expect } from "vitest";

import { legalActionsOf, idleReadyHint, readyPhrases } from "./legalActions";
import type { CardView, GameView, LegalMoveView, LegalSourceView } from "./protocol";

const HINT = "Overloaded, this does nothing right now: there's no spell you don't control.";
const FLUX = "11111111-1111-1111-1111-111111111111";

function digestView(entry: LegalSourceView): GameView {
  return { legal_actions: { pass: true, sources: { [FLUX]: entry } } } as unknown as GameView;
}

function cast(alt: string, hint?: string): LegalMoveView {
  return {
    type: "cast_spell",
    player: "p",
    kind: "cast",
    label: "Cast Counterflux",
    source: FLUX,
    params: { instance_id: FLUX, from_zone: "hand", alternative_cost: alt },
    ...(hint ? { idle_hint: hint } : {}),
  };
}

function movesView(moves: LegalMoveView[]): GameView {
  const pass: LegalMoveView = { type: "pass_priority", player: "p", kind: "pass", label: "Pass" };
  return { legal_moves: [pass, ...moves] } as unknown as GameView;
}

describe("idleReadyHint (#1918)", () => {
  it("reads the digest's cast_idle_hint: every cast idle → muted, with the hint", () => {
    const legal = legalActionsOf(
      digestView({ kinds: ["cast"], moves: 1, zones: ["hand"], cast_idle_hint: HINT }),
    );
    expect(legal.castableFrom(FLUX, "hand")).toBe(true);
    expect(idleReadyHint(legal, FLUX, "hand")).toBe(HINT);
  });

  it("no cast_idle_hint in the digest → a normal ring", () => {
    const legal = legalActionsOf(digestView({ kinds: ["cast"], moves: 2, zones: ["hand"] }));
    expect(legal.castableFrom(FLUX, "hand")).toBe(true);
    expect(idleReadyHint(legal, FLUX, "hand")).toBeUndefined();
  });

  it("from legal_moves: every cast hinted → muted", () => {
    const legal = legalActionsOf(movesView([cast("overload", HINT)]));
    expect(idleReadyHint(legal, FLUX, "hand")).toBe(HINT);
  });

  it("from legal_moves: a mix of idle and live casts → a normal ring, in either order", () => {
    for (const moves of [
      [cast("", undefined), cast("overload", HINT)],
      [cast("overload", HINT), cast("", undefined)],
    ]) {
      const legal = legalActionsOf(movesView(moves));
      expect(legal.castableFrom(FLUX, "hand")).toBe(true);
      expect(idleReadyHint(legal, FLUX, "hand")).toBeUndefined();
    }
  });

  it("a card ready for something else too (an ability, a special action) is not muted", () => {
    const extras: Partial<LegalSourceView>[] = [
      { abilities: ["own:0"], kinds: ["cast", "activate"] },
      { special_actions: ["foretell"], kinds: ["cast", "special_action"] },
    ];
    for (const extra of extras) {
      const legal = legalActionsOf(
        digestView({ kinds: [], moves: 2, zones: ["hand"], cast_idle_hint: HINT, ...extra }),
      );
      expect(idleReadyHint(legal, FLUX, "hand")).toBeUndefined();
    }
  });

  it("the accessible name says why the ring is muted", () => {
    const card = { instance_id: FLUX, name: "Counterflux" } as CardView;
    const idle = legalActionsOf(
      digestView({ kinds: ["cast"], moves: 1, zones: ["hand"], cast_idle_hint: HINT }),
    );
    expect(readyPhrases(idle, card, "hand")).toEqual([
      "castable, but overloaded, this does nothing right now: there's no spell you don't control",
    ]);
    const live = legalActionsOf(digestView({ kinds: ["cast"], moves: 1, zones: ["hand"] }));
    expect(readyPhrases(live, card, "hand")).toEqual(["castable"]);
  });

  it("not castable from this zone → no hint", () => {
    const legal = legalActionsOf(
      digestView({ kinds: ["cast"], moves: 1, zones: ["hand"], cast_idle_hint: HINT }),
    );
    expect(idleReadyHint(legal, FLUX, "graveyard")).toBeUndefined();
  });
});
