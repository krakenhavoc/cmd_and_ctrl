package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b17_graveyard.go — the seam "was put into a graveyard this turn"
// (Abyssal Harvester, #4465).

// PutIntoAGraveyardThisTurn passes a card that was put into a graveyard
// from anywhere this turn and is still there: it died, was discarded, was
// milled or was put there by an effect. It reads Card.PutInGraveyardThisTurn,
// the per-turn flag the zone move stamps on every arrival in a graveyard and
// clears when the card leaves it and at the next turn (ADR 0066's 2026-10-15
// amendment), so a card that came back out and went in again counts and one
// whose arrival was an earlier turn's does not.
func PutIntoAGraveyardThisTurn() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.PutInGraveyardThisTurn
	}
}

// tokenCopyIsAlsoASubtype is "except it's a <subtype> in addition to its other
// types" on a token copy's template (CR 707.9b): the subtype, read through the
// text of the object that made the copy, joins the copiable values.
func tokenCopyIsAlsoASubtype(t *game.Card, subtype string) {
	v := game.PrintedValues{TypeLine: t.TypeLine}
	v.AddSubtype(subtype)
	t.TypeLine = v.TypeLine
}
