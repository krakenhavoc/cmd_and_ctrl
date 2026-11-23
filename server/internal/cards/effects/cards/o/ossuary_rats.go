package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ossuary Rats — {5}{B} Creature — Rat 3/2 (Jumpstart 2022, #5948):
//
//	"When this creature enters, it deals X damage to target creature or
//	 planeswalker an opponent controls, where X is the number of creature
//	 cards in your graveyard."
//
// X is counted as the trigger resolves (CR 608.2h), after the target is
// re-checked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1380d07-f582-46dd-a63a-9806123ebcd3",
		Name:         "Ossuary Rats",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ossuary Rats — damage to target creature or planeswalker an opponent controls equal to the creature cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: effects.B11CreatureCardsInGraveyard(g, item.Controller)}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetPermanent("target creature or planeswalker an opponent controls",
					effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()))),
		},
	})
}
