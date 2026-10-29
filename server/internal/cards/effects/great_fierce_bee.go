package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Great Fierce Bee — Creature — Insect {2}{B}, 2/2 (The Hobbit):
//
//	"Flying
//	 Whenever one or more other creatures die, scry 1."
//
// "One or more" is a batch (CR 603.2c, 603.2d): a wipe that kills five
// creatures is one scry, not five (OncePerBatch). Any creature counts,
// yours or an opponent's, and not the Bee itself: a wipe that takes the
// Bee along with the rest still triggers for the others, because they
// died in the same batch (CR 603.10a), and the Bee dying alone does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "37f82ad8-4f0b-4b17-abcc-03f43d4fdf1c",
		Name:            "Great Fierce Bee",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, AnotherCreatureDied, "Great Fierce Bee — scry 1", Do(Scry{N: 1}))),
		},
	})
}
