package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashino Sandstalker — Creature — Lizard Warrior {1}{R}{R}, 4/2
// (Multiverse Gift Box, #5632):
//
//	"Haste
//	 At the beginning of the end step, return this creature to its
//	 owner's hand."
//
// "At the beginning of the end step" is ANY end step, not only yours
// (AtEachStep, Lightning Serpent's shape), so one flashed in or put onto
// the battlefield on another player's turn goes back at that turn's end.
// It returns only if it is still on the battlefield as the same object
// (BounceToHand's CR 400.7 guard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cbfccfc0-3117-4ca0-b7ce-8e2d67ec49ab",
		Name:            "Viashino Sandstalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEnd, "Viashino Sandstalker — return it to its owner's hand", func(g *game.Game, item *game.StackItem) error {
				return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
	})
}
