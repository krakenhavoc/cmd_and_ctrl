package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace, Ingenious Mind-Mage — Legendary Planeswalker — Jace {4}{U}{U},
// loyalty 5 (Ixalan, #5532):
//
//	"+1: Draw a card.
//	 +1: Untap all creatures you control.
//	 −9: Gain control of up to three target creatures."
//
// Two separate +1 abilities, either of which may be activated (once a turn
// in all, CR 606.3). The untap is every creature you control as it
// resolves. The −9 takes up to three targets, none being fine, and the
// control has no duration: it lasts until the game ends or the creature
// leaves (CR 611.2b); a target no longer legal as it resolves is not taken
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe0fcb4e-92fd-4e64-99d1-043092ea21e3",
		Name:            "Jace, Ingenious Mind-Mage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "+1: Untap all creatures you control.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{},
				Effect:  effects.Do(effects.UntapAllCreaturesYouControl{}),
			},
			{
				Label:   "−9: Gain control of up to three target creatures.",
				Cost:    effects.LoyaltyCost(-9),
				Targets: effects.TargetCreature("up to three target creatures").WithCount(0, 3),
				Purpose: game.Purpose{},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.GainControl{
							Target:     t.ID,
							Controller: item.Controller,
							Duration:   game.IndefiniteDuration(),
							Label:      "Jace, Ingenious Mind-Mage — gain control (no stated duration)",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
