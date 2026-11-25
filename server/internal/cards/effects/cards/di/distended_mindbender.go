package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Distended Mindbender — Creature — Eldrazi Insect {8}, 5/5 (Eldritch Moon,
// S82 OLD, #5541):
//
//	"Emerge {5}{B}{B} (You may cast this spell by sacrificing a creature and
//	 paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, target opponent reveals their hand. You
//	 choose from it a nonland card with mana value 3 or less and a card with
//	 mana value 4 or greater. That player discards those cards."
//
// Emerge is the shared alternative cost (ADR 0135 §4). The cast trigger is
// Mockery of Nature's (WhenYouCastThisSpell: it resolves above the spell
// even if it is countered) with the one-pick-of-two revealed-hand choice
// (#6271, ChooseFromRevealedHand.EachOf): one card for each filter, each a
// different card, from one reveal (CR 701.20a); a hand that can fill only
// one filter gives up that one, and the cards are discarded together
// (CR 701.9a).
//
// No simplification.
func init() {
	cast := effects.WhenYouCastThisSpell("Distended Mindbender — target opponent reveals their hand; you choose a nonland card with mana value 3 or less and a card with mana value 4 or greater; they discard those cards",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				EachOf: []effects.CardPredicate{
					effects.And(effects.Nonland(), effects.ManaValueLE(3)),
					effects.ManaValueGE(4),
				},
				Label: "a nonland card with mana value 3 or less and a card with mana value 4 or greater",
			}.Apply(ctx)
		})
	cast.Targets = effects.TargetPlayer("target opponent", effects.Opponent())
	effects.Register(effects.Spec{
		OracleID:         "36d32194-d547-4e2d-894f-f4bf13cf7da6",
		Name:             "Distended Mindbender",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Emerge("{5}{B}{B}")},
		Triggered:        []game.TriggeredAbility{cast},
	})
}
