// engineMayMissMana.ts — "stop if the engine may be wrong" (ADR 0118,
// owner decision 8, 2026-10-04).
//
// With strict payment on by default, the hand is dimmed by the server's
// move list, which is strict and auto-tapped: a card the ENGINE cannot
// pay for gets no cast move. Smart autopass reads the same list, so on
// the viewer's own main phase it finds nothing to play and passes the
// step. That is right when the engine sees every mana source the
// player has. It is wrong when the player controls a permanent whose
// mana the engine does not make: the card IS payable, the engine just
// cannot see how, and "Cast anyway (don't pay)" exists for exactly that
// case. A passed main phase would hide it.
//
// So smart autopass holds the viewer's own main phase when BOTH:
//
//  1. the viewer holds a spell the server's move list does not offer,
//     for no reason but mana: the hand card has a castable face, the
//     move list has no cast for it, and castAnywayBlocked (the Cast
//     anyway row's own non-mana denials) says nothing; and
//  2. the viewer controls a manual mana source, read from the wire as
//     a permanent with `unimplemented` set (it prints rules the engine
//     will not run) that either HAS a mana ability on the wire
//     (`mana_abilities`) or is PRESUMED to have one because it is a
//     land.
//
// What the wire cannot tell: whether an unimplemented NON-land
// permanent with no `mana_abilities` row (an uncatalogued mana creature
// or mana rock) makes mana. `unimplemented` says the card prints rules
// the engine will not run, not which rules, and CardView carries no
// oracle text. Those permanents do not trigger the stop; telling them
// apart needs a server field.
//
// Pure functions over the snapshot. The gate is assembled in
// Game.svelte and read by autopassDecision.ts's stops-grid rule.

import { isLand } from "./cardTypes";
import type { CardView, GameView } from "./protocol";
import {
  castAnywayBlocked,
  castAnywayOffered,
  isActivePlayer,
  isMainPhase,
  movesFor,
} from "./timing";

// unpayableSpellInHand reports whether the viewer holds a spell the
// server's move list leaves out for lack of mana alone (condition 1).
// No move list on the frame means no information, so it says no.
export function unpayableSpellInHand(
  view: GameView | null | undefined,
  me: string | null,
): boolean {
  if (!view || !me || !view.legal_moves) return false;
  const hand = view.seats?.find((s) => s.id === me)?.hand?.cards ?? [];
  return hand.some((c) => {
    if (!castAnywayOffered(c, "hand")) return false;
    const casts = movesFor(view, c.instance_id, ["cast"]);
    if (!casts || casts.length > 0) return false;
    return castAnywayBlocked(c, view, me, "hand") === "";
  });
}

// isManualManaSource is condition 2 for one permanent: the engine does
// not run all of its printed rules, and it has, or as a land is
// presumed to have, a mana ability.
export function isManualManaSource(c: CardView): boolean {
  if (!c.unimplemented) return false;
  return (c.mana_abilities?.length ?? 0) > 0 || isLand(c);
}

// controlsManualManaSource reports whether the viewer controls a
// manual mana source on the battlefield (condition 2).
export function controlsManualManaSource(
  view: GameView | null | undefined,
  me: string | null,
): boolean {
  if (!view || !me) return false;
  return (view.battlefield?.cards ?? []).some((c) => c.controller === me && isManualManaSource(c));
}

// engineMayMissMana is the whole gate: the viewer's own main phase, and
// both conditions. Anywhere else it is false and autopass is unchanged.
export function engineMayMissMana(view: GameView | null | undefined, me: string | null): boolean {
  if (!view || !me) return false;
  if (!isMainPhase(view) || !isActivePlayer(view, me)) return false;
  return unpayableSpellInHand(view, me) && controlsManualManaSource(view, me);
}
