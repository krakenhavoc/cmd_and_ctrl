package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Root Elemental — Creature — Elemental {4}{G}{G}, 6/5:
//
//	"Morph {5}{G}{G}
//	 When this creature is turned face up, you may put a creature card
//	 from your hand onto the battlefield."
//
// Champion of Rhonas's put: the card is put, not cast, and the "may" is
// a real decline.
//
// No simplification.
func init() {
	const label = "Root Elemental — you may put a creature card from your hand onto the battlefield"
	effects.Register(effects.Spec{
		OracleID:         "182d38b4-48bb-4050-a19a-c0a6bb241bbd",
		Name:             "Root Elemental",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{5}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp(label, func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{Match: effects.Creature(), Optional: true, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
