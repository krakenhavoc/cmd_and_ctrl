package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raging Kronch — Creature — Beast {2}{R}, 4/3 (War of the Spark, #5520):
//
//	"This creature can't attack alone."
//
// Pet Sehlat's attack limit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c05c2eae-1a9c-4ce7-9c76-22e40781c1a9",
		Name:         "Raging Kronch",
		Completeness: effects.CompletenessFull,
		AttackLimits: []game.AttackLimit{effects.CantAttackAlone()},
	})
}
