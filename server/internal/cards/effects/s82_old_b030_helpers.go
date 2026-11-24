package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b030_helpers.go — shared pieces of OLD batch 30 (#5501, Modern
// Horizons 2).

// OldB030DiscardedACardThisTurn is "you've discarded a card this turn": the
// controller's discard tally for the turn (CR 701.9a), one per card.
func OldB030DiscardedACardThisTurn(g *game.Game, player uuid.UUID) bool {
	return g.TurnTallyFor(player).CardsDiscarded > 0
}

// OldB030SacrificeUnlessDiscard is "sacrifice this permanent unless you
// discard a card" at an upkeep (Solitary Confinement): the discard is asked
// as a choice of up to one card, and declining it, or holding no card,
// sacrifices the permanent if it is still the object the trigger came from.
func OldB030SacrificeUnlessDiscard(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Eliminated {
			return nil
		}
		if p.Hand == nil || p.Hand.Size() == 0 {
			return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
		}
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   item.Controller,
			Source:   item.SourceCardID,
			N:        1,
			UpTo:     true,
			Question: name + " — discard a card, or sacrifice it?",
			Asked:    game.NewAskedItem(item),
		}, OldB030SacrificedUnlessDiscarded, nil)
	}
}

// OldB030SacrificedUnlessDiscarded is the discard answered: nothing
// discarded sacrifices the permanent (CR 400.7: only if it is still the
// object that asked). An on-disk key: never renamed, never reused.
var OldB030SacrificedUnlessDiscarded = game.RegisterDiscardThen("choose-cards/old-b030-sacrifice-unless-discard",
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Item == nil || len(r.Discarded.Cards()) > 0 || !SourceIsStillThisPermanent(g, r.Item) {
			return nil
		}
		return SacrificePermanent{Target: r.Item.SourceCardID}.Apply(NewContext(g, r.Item))
	})
