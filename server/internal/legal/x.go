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
// ONE GAP LEFT, and it is "the enumerator is not choosing X here":
//
//   - CountFromX on an ACTIVATED ability. The engine resolves an
//     X-defined target count for casts only (game/mutations.go), so
//     the tie in cast.go has no ability-side twin to mirror.
//
// The other one — X announced by a cost that is not the mana cost —
// is closed for the "pay X life" half by xCeilingFromCost below
// (#957). Waterbender's Restoration's waterbend TapPermanentsCost is
// still unpriced; cast.go declines to enumerate that cast rather than
// announcing an X it cannot pay for.
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

// noXCeiling means "no cost component other than the mana cost prices
// X", which is every card but the pay-X-life family today.
const noXCeiling = -1

// xCeilingFromCost is the largest X a NON-MANA cost component lets the
// seat announce, or noXCeiling when no component prices one (#957).
//
// X is ONE announced number (CR 601.2b) and the mana cost is not the
// only thing that can charge for it. Toxic Deluge prints {2}{B} with
// no {X} anywhere in it and announces X by paying X life, so the
// {X}-slot search in cast.go answers 0 for it and always would — the
// spell was offered at X=0 and swept the board for -0/-0, #810's
// zero-effect move arriving through a different seam.
//
// Keyed on the COST COMPONENT, not on the card: any future "as an
// additional cost, pay X life" is priced by this line the day it is
// registered, with no catalog entry and no per-card branch. The floor
// is still enumeratedXFloor's — Toxic Deluge declares XMatters, so it
// is 1 and a seat that cannot reach 1 is offered no cast at all.
//
// The ceiling is life - 1, not life. CR 119.4 permits paying exactly
// your life total, and the engine accepts it (the announce check in
// game/additional_cost.go is `xValue > p.Life`); this package will not
// OFFER it, on the same "what is worth putting in front of a player"
// footing as the X=0 rule above. A sweep that kills the caster is not
// a move a bot should be handed as its only pricing of the card.
func xCeilingFromCost(addCost *game.AdditionalCost, life int) int {
	if addCost == nil || !addCost.PayLifeX {
		return noXCeiling
	}
	if life < 1 {
		return 0
	}
	return life - 1
}

// announcedX is the one number a cast announces for X: the largest
// value EVERY cost component that prices X can pay for, at or above
// the floor, capped by Options.MaxX like any other X search. Reports
// false when the floor cannot be met — no move at all, rather than a
// free one.
//
// `priced` is the mana cost after modifiers; `lifeCeiling` comes from
// xCeilingFromCost. A cost with no {X} slot has no mana ceiling, so
// the life ceiling is the whole answer (Toxic Deluge); a cost with
// both would take the smaller, and a cost with neither keeps
// affordableXFrom's answer unchanged.
func (e *enumerator) announcedX(
	priced game.ParsedCost,
	spend game.ManaSpendContext,
	floor int,
	lifeCeiling int,
) (int, bool) {
	x, ok := e.affordableXFrom(priced, spend, floor)
	if !ok {
		return 0, false
	}
	if lifeCeiling == noXCeiling {
		return x, true
	}
	if priced.XSlots == 0 || lifeCeiling < x {
		x = lifeCeiling
	}
	if x > e.opts.MaxX {
		x = e.opts.MaxX
	}
	if x < floor {
		return 0, false
	}
	return x, true
}
