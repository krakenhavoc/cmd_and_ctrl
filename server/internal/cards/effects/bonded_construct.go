package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bonded Construct — Artifact Creature — Construct {1}, 2/1 (Magic Origins, #5549):
//
//	"This creature can't attack alone."
//
// Pet Sehlat's attack limit (CR 506.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7556be3-649a-4994-8aaa-a0f258856682",
		Name:         "Bonded Construct",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{CantAttackAlone()},
	})
}
