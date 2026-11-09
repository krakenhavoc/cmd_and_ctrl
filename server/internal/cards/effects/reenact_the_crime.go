package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reenact the Crime — Instant {1}{U}{U}{U} (Murders at Karlov Manor, #4469):
//
//	"Exile target nonland card in a graveyard that was put there from
//	 anywhere this turn. Copy it. You may cast the copy without paying its
//	 mana cost."
//
// The target is a card in any graveyard whose arrival record says it was put
// there this turn, from any zone (Card.PutInGraveyardThisTurn). It is exiled
// first, and the copy is made of the card that really reached exile (CR
// 707.12, CopyAndCast): the cast is made as part of this resolution (CR
// 608.2g), free (X at 0, no alternative cost), and a copy not cast ceases to
// exist. A target that is no longer legal makes the spell do nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebb88893-427f-496c-82a8-94d2bafb5e4c",
		Name:         "Reenact the Crime",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target nonland card in a graveyard that was put there from anywhere this turn",
			Nonland(), PutIntoAGraveyardThisTurn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ExileTarget{Target: id, ThenTail: ctx.ItemTail(reenactTheCrimeCopy)}.Apply(ctx)
		},
	})
}

// reenactTheCrimeCopy is "copy it, you may cast the copy without paying its
// mana cost", told the card that reached exile.
var reenactTheCrimeCopy = CardTail("tail/reenact-the-crime-copy", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CopyAndCast{Cards: o.IDs, Free: true}.Apply(ctx)
})

// PutIntoAGraveyardThisTurn is "that was put there from anywhere this turn":
// the card is in a graveyard and arrived during the current turn, from any
// zone (Card.PutInGraveyardThisTurn).
func PutIntoAGraveyardThisTurn() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.PutInGraveyardThisTurn }
}
