package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master of Pearls — Creature — Human Monk {1}{W}, 2/2:
//
//	"Morph {3}{W}{W}
//	 When this creature is turned face up, creatures you control get
//	 +2/+2 until end of turn."
//
// The creatures are locked in as the trigger resolves (CR 611.2c).
//
// No simplification.
func init() {
	const label = "Master of Pearls — creatures you control get +2/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:         "aca9df28-2df6-477c-b5b8-8e1863c84356",
		Name:             "Master of Pearls",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{3}{W}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp(label, func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 2, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
