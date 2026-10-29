package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stony-Voiced Goblins — Creature — Goblin Bard {1}{B}, 1/1 (The Hobbit):
//
//	"When this creature enters, each opponent discards a card."
//
// The same rules text as Burglar Rat and Rank Rat: one discard prompt
// per opponent, each choosing their own card (eachOpponentDiscardsOne).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1dceaa7-d33e-4067-9e1b-8a8978a12933",
		Name:         "Stony-Voiced Goblins",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stony-Voiced Goblins — each opponent discards a card", eachOpponentDiscardsOne),
		},
	})
}
