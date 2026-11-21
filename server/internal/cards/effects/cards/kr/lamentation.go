package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lamentation — Creature — Elemental Incarnation {5}{B}, 5/4 (Lorwyn
// Eclipsed Commander):
//
//	"When this creature enters, destroy target creature an opponent
//	 controls. You gain 3 life.
//	 Encore {6}{B}{B} (...)"
//
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b); with no legal target left the whole ability is
// removed from the stack and no life is gained. Encore is the shared keyword
// (encore.go, CR 702.141a): each token copy has the enters trigger too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88e43e62-8bce-41b3-ac2d-47043fcc7077",
		Name:         "Lamentation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Lamentation — destroy target creature an opponent controls. You gain 3 life",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.DestroyTheFirstLegalTarget(g, item); err != nil {
						return err
					}
					return effects.GainLife{Amount: 3}.Apply(effects.NewContext(g, item))
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{effects.Encore("{6}{B}{B}")},
	})
}
