package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindstorm Crown — Artifact {3} (Mirrodin, S82 OLD, #5606):
//
//	"At the beginning of your upkeep, draw a card if you had no cards in
//	 hand at the beginning of this turn. If you had a card in hand, this
//	 artifact deals 1 damage to you."
//
// Two sentences, one trigger that always goes on the stack (neither is an
// intervening "if", CR 603.4). The hand is counted as the upkeep begins,
// which is the hand the turn began with (nothing can change it in the
// untap step, CR 502.4), and carried on the item (Params.Amount) so a card
// cast or drawn in response cannot change the answer. With an empty hand
// you draw; with any card in it, the Crown deals 1 damage to you (a source
// that is an artifact, CR 120.3).
//
// No simplification.
func init() {
	const label = "Mindstorm Crown — draw a card if you had no cards in hand at the beginning of this turn; otherwise it deals 1 damage to you"
	t := effects.AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		if item.Params.Amount == 0 {
			return effects.DrawCards{N: 1}.Apply(ctx)
		}
		return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(ctx)
	})
	t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
			item.Params.Amount = p.Hand.Size()
		}
		return item
	}
	effects.Register(effects.Spec{
		OracleID:     "0ad26c65-bc2c-4aa7-8c5d-e4c8f0629cb7",
		Name:         "Mindstorm Crown",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
