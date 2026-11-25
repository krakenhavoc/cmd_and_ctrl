package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basilica Bell-Haunt — Creature — Spirit {W}{W}{B}{B}, 3/4 (Ravnica
// Allegiance):
//
//	"When this creature enters, each opponent discards a card and you gain 3
//	 life."
//
// Each opponent chooses their own discard (CR 701.9a); the life is gained
// whatever they do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "796dc53d-5a0d-4d19-9fa4-f474a9d96041",
		Name:         "Basilica Bell-Haunt",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Basilica Bell-Haunt — each opponent discards a card and you gain 3 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					effects.EachOpponentDiscards(ctx, 1)
					return effects.GainLife{Amount: 3}.Apply(ctx)
				}),
		},
	})
}
