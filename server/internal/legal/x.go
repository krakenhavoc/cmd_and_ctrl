package legal

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// x.go — the enumerator's rule for the announced X (#810, #619).
//
// ONE rule, shared by casts (cast.go) and activations (abilities.go),
// because the mistake it fixes was the same on both sides and a second
// copy is how they drift apart.
//
// The rule is about what to OFFER, not about what is legal. CR 601.2b
// and 602.2b make X a number the caster announces, and where the
// printed text puts no floor on it, 0 is a legal announcement: the
// engine accepts a Soothsaying activation at X=0 and is right to. But
// "{X}: Look at the top X cards of your library" at X=0 costs nothing
// and does nothing, and the identical move is back on the list the
// instant it resolves — a repeatable sequence of optional actions
// that changes nothing, which CR 732.2a says no player is ever made
// to keep repeating. A table of bots took that offer 79,519 times in
// five minutes and never finished turn 18 (#810).
//
// So: a move whose whole effect is X is not offered at X=0. The
// minimum this package announces for such a cost is 1, and when X=1
// cannot be paid the move is not offered at all — Soothsaying with an
// empty board is not a move, the way Helm of Obedience's printed "X
// can't be 0" already was not one.
//
// WHAT COUNTS AS "whose whole effect is X" is the catalog's to say,
// not this package's to guess: `effects.Spec.XMatters`, read here
// through game.XMattersFor, and kept honest by the source scan in
// effects/x_matters_guard_test.go. A card with a fixed RIDER — The
// Goose Mother's 2/2 flying body, Springleaf Parade's mana static —
// leaves it unset, and X=0 stays on the list for it. That is the
// other half of the rule and it costs nothing to state, because the
// search below takes the LARGEST affordable X: a rider's X=0 is
// offered exactly when no larger announcement can be paid for.
//
// TWO GAPS, both of them "the enumerator is not choosing X here":
//
//   - X from a cost that is not the mana cost. Toxic Deluge announces
//     X by paying X life (AdditionalCost.PayLifeX) and Waterbender's
//     Restoration by a waterbend TapPermanentsCost. Nothing in this
//     package prices either, so it announces 0 and the rule below
//     never fires (the floor is meaningless on a cost with no {X}
//     slot). Toxic Deluge is therefore still offered at X=0.
//   - CountFromX on an ACTIVATED ability. The engine resolves an
//     X-defined target count for casts only (game/mutations.go), so
//     the tie in cast.go has no ability-side twin to mirror.
func enumeratedXFloor(catalogKey string, printedFloor int) int {
	// The printed floor wins when it is higher: Helm of Obedience's
	// MinX(1) is a rule of the card, this is a rule about offers, and
	// a card that ever prints "X can't be 2" would be announced at 3.
	if printedFloor >= 1 {
		return printedFloor
	}
	if game.XMattersFor(catalogKey) {
		return 1
	}
	return printedFloor
}
