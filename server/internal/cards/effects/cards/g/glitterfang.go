package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glitterfang — Creature — Spirit {R}, 1/1 (Saviors of Kamigawa, #5601):
//
//	"Haste
//	 At the beginning of the end step, return this creature to its
//	 owner's hand."
//
// Viashino Sandstalker's shape: "the end step" is every end step, not
// only yours (AtEachStep), and the return goes through BounceToHand's
// CR 400.7 guard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "226dca3d-9da3-4c3f-89d0-67264fd19c51",
		Name:            "Glitterfang",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Glitterfang — return it to its owner's hand", func(g *game.Game, item *game.StackItem) error {
				return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
