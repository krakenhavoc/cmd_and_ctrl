package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_x_cast_cost.go — "As an additional cost to cast this spell,
// discard X cards" (#6145, game.AdditionalCost.DiscardX). X has no {X}
// slot to price it, so, as for "exile X cards from your graveyard"
// (exile_card_cast_cost.go), the hand is the ceiling on the X a move
// announces and the payment is the first X cards of the pool.

// maxEnumeratedDiscardX caps the X a "discard X cards" move announces.
// Any X up to the hand is legal; the enumerator offers a small one so a
// bot never pitches its whole hand to a two-damage Channeled Force.
const maxEnumeratedDiscardX = 2

// discardXCeiling is the largest X a "discard X cards" cost lets the seat
// announce in a move: the engine's own DiscardXCeilingForEffect (the
// number CastSpell refuses above, #544), capped at maxEnumeratedDiscardX.
// noXCeiling for every other cost.
func (e *enumerator) discardXCeiling(castID uuid.UUID, addCost *game.AdditionalCost) int {
	if addCost == nil || !addCost.DiscardX {
		return noXCeiling
	}
	return min(e.g.DiscardXCeilingForEffect(e.seat, castID, addCost), maxEnumeratedDiscardX)
}

// discardCardsForX is the discard a move at X names for a "discard X
// cards" cost: the first X of the pool. Nil for every other cost.
func discardCardsForX(paying *game.AdditionalCost, pool []uuid.UUID, x int) []uuid.UUID {
	if paying == nil || !paying.DiscardX || x <= 0 {
		return nil
	}
	return pool[:min(x, len(pool))]
}

// maxEnumeratedReturnX caps the X a "return X <permanents> you control"
// move announces (#6122, Infernal Harvest): any X up to the board is
// legal; the enumerator offers a small one so a bot never bounces every
// land it has.
const maxEnumeratedReturnX = 2

// returnXCeiling is the largest X a "return X …" additional cost lets the
// seat announce in a move: the engine's own ReturnToHandXCeilingForEffect,
// capped at maxEnumeratedReturnX. noXCeiling for every other cost.
func (e *enumerator) returnXCeiling(castID uuid.UUID, addCost *game.AdditionalCost) int {
	if addCost == nil || !game.ReturnToHandCountFromX(addCost.ReturnToHand) {
		return noXCeiling
	}
	return min(e.g.ReturnToHandXCeilingForEffect(e.seat, castID, addCost.ReturnToHand), maxEnumeratedReturnX)
}

// returnPermanentsForX is the return a move at X names for a "return X …"
// cost: the first X of the pool. Nil for every other cost.
func returnPermanentsForX(paying *game.AdditionalCost, pool []uuid.UUID, x int) []uuid.UUID {
	if paying == nil || !game.ReturnToHandCountFromX(paying.ReturnToHand) || x <= 0 {
		return nil
	}
	return pool[:min(x, len(pool))]
}
