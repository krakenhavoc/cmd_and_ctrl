package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonded Horncrest — Creature — Dinosaur {3}{R}, 5/5 (Ixalan, #5532):
//
//	"This creature can't attack or block alone."
//
// Mogg Flunkies' rules: it can't be declared as the only attacker
// (CR 506.5) or as the only blocker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf71c210-8fef-4eda-a943-8c52cbae6159",
		Name:         "Bonded Horncrest",
		Completeness: effects.CompletenessFull,
		AttackLimits: []game.AttackLimit{effects.CantAttackAlone()},
		BlockRules:   []game.BlockRule{effects.CantBlockAlone()},
	})
}
