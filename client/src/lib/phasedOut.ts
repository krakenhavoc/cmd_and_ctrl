// phasedOut.ts — the client's half of CR 702.26 (#1199, ADR 0084).
//
// A phased-out permanent is "treated as though it does not exist"
// (CR 702.26b), and the SERVER makes that true: it is not in
// `GameView.battlefield` at all, it is in `GameView.phased_out`, and
// nothing that reads the battlefield — targeting, legal moves, the
// bot — can see it. The client is the one consumer that deliberately
// looks at the other zone, because a board that silently loses four
// permanents is indistinguishable from a board that was wrathed.
//
// So the rendering rule is: show it WHERE IT WAS, dimmed, badged, and
// inert. It reads as "still yours, currently not here", which is what
// the card says.
//
// A pure helper rather than logic inside Board.svelte, for
// cardBack.ts's reason (client/README.md): the decision is testable
// without mounting anything.

import type { CardView, GameView } from "./protocol";

// isPhasedOut reports whether a CardView is a phased-out permanent.
//
// The `!== undefined` / `=== true` shape rather than a truthiness test
// is the house convention for an `omitempty` boolean: the field is
// ABSENT on the wire when false, never `false`.
export function isPhasedOut(card: CardView): boolean {
  return card.phased_out === true;
}

// phasedOutCards returns the phased-out permanents on a view, or an
// empty list. Tolerates a view from a server that predates the zone.
export function phasedOutCards(view: GameView): CardView[] {
  return view.phased_out?.cards ?? [];
}

// phasedOutByController groups them the way the board groups the
// battlefield — by CONTROLLER, not owner.
//
// Controller is right and owner is not: CR 702.26d says the phasing
// event doesn't change control, so a creature stolen with Act of
// Treason and then phased out is still on the thief's side of the
// table, and that is where the player expects to see it come back.
export function phasedOutByController(view: GameView): Map<string, CardView[]> {
  const out = new Map<string, CardView[]>();
  for (const c of phasedOutCards(view)) {
    const list = out.get(c.controller);
    if (list) list.push(c);
    else out.set(c.controller, [c]);
  }
  return out;
}
