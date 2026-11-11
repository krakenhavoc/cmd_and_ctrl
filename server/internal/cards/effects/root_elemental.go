package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "182d38b4-48bb-4050-a19a-c0a6bb241bbd",
		Name:             "Root Elemental",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{5}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp(label, func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{Match: Creature(), Optional: true, Label: label}.Apply(NewContext(g, item))
			}),
		},
	})
}
