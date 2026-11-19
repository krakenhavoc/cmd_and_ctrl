package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mogg Flunkies — Creature — Goblin {1}{R}, 3/3 (Foundations Jumpstart):
//
//	"This creature can't attack or block alone."
//
// Both halves are restrictions on the creature itself (CR 506.5, 508.1c,
// 509.1b): the attack limit CantAttackAlone and the block rule
// CantBlockAlone, as Toby's Beast token carries them. Another creature
// that can't attack or block alone may join it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b92fc9c1-7264-4894-90b8-a193039489c5",
		Name:         "Mogg Flunkies",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{CantAttackAlone()},
		BlockRules:   []game.BlockRule{CantBlockAlone()},
	})
}
