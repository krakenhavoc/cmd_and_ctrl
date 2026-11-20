package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bonded Horncrest — Creature — Dinosaur {3}{R}, 5/5 (Ixalan, #5532):
//
//	"This creature can't attack or block alone."
//
// Mogg Flunkies' rules: it can't be declared as the only attacker
// (CR 506.5) or as the only blocker.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf71c210-8fef-4eda-a943-8c52cbae6159",
		Name:         "Bonded Horncrest",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{CantAttackAlone()},
		BlockRules:   []game.BlockRule{CantBlockAlone()},
	})
}
