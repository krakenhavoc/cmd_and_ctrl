package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jackal Familiar — Creature — Jackal {R}, 2/2 (Magic 2010, #5585):
//
//	"This creature can't attack or block alone."
//
// Mogg Flunkies' two restrictions: it can attack only beside another
// attacker and block only beside another blocker.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4d5d666-fd55-4c1a-89c8-2348288f6b30",
		Name:         "Jackal Familiar",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{CantAttackAlone()},
		BlockRules:   []game.BlockRule{CantBlockAlone()},
	})
}
