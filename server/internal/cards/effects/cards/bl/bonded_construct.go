package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonded Construct — Artifact Creature — Construct {1}, 2/1 (Magic Origins, #5549):
//
//	"This creature can't attack alone."
//
// Pet Sehlat's attack limit (CR 506.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7556be3-649a-4994-8aaa-a0f258856682",
		Name:         "Bonded Construct",
		Completeness: effects.CompletenessFull,
		AttackLimits: []game.AttackLimit{effects.CantAttackAlone()},
	})
}
