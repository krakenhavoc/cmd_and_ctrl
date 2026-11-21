package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ember Beast — Creature — Beast {2}{R}, 3/4 (Dominaria Remastered, #5948):
//
//	"This creature can't attack or block alone."
//
// Both halves are restrictions on the creature itself (CR 506.5, 508.1c,
// 509.1b): CantAttackAlone and CantBlockAlone, as Mogg Flunkies carries
// them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58edb282-a124-47bc-b123-1c34b6804f48",
		Name:         "Ember Beast",
		Completeness: effects.CompletenessFull,
		AttackLimits: []game.AttackLimit{effects.CantAttackAlone()},
		BlockRules:   []game.BlockRule{effects.CantBlockAlone()},
	})
}
