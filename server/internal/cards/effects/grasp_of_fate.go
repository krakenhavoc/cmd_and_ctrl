package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grasp of Fate — Enchantment {1}{W}{W}:
//
//	"When this enchantment enters, for each opponent, exile up to one
//	 target nonland permanent that player controls until this
//	 enchantment leaves the battlefield."
//
// One up-to-one clause per opponent, each bound to that opponent
// (perOpponentClauses). An enchantment that has already left by
// resolution exiles nothing (CR 610.3c); the exiled cards return as it
// leaves, tokens ceasing to exist.
func init() {
	exile := WhenThisEnters("Grasp of Fate — exile a nonland permanent of each opponent until it leaves", cmdrB1GraspExileEach)
	exile.Targets = perOpponentClauses(game.MaxPlayers-1, "nonland permanent", Nonland())
	Register(Spec{
		OracleID:     "1b26a54f-6c27-4940-b77e-2a66175c5eff",
		Name:         "Grasp of Fate",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{exile},
	})
}
