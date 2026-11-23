package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	exile := effects.WhenThisEnters("Grasp of Fate — exile a nonland permanent of each opponent until it leaves", effects.CmdrB1GraspExileEach)
	exile.Targets = effects.PerOpponentClauses(game.MaxPlayers-1, "nonland permanent", effects.Nonland())
	effects.Register(effects.Spec{
		OracleID:     "1b26a54f-6c27-4940-b77e-2a66175c5eff",
		Name:         "Grasp of Fate",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{exile},
	})
}
