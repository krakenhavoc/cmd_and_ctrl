package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Okk — Creature — Goblin {1}{R}, 4/4 (Eighth Edition, #5607):
//
//	"This creature can't attack unless a creature with greater power also
//	 attacks.
//	 This creature can't block unless a creature with greater power also
//	 blocks."
//
// Both limits compare power as the declaration is made (CR 208.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f835a8d3-e117-4669-9952-6e9ce13b35fa",
		Name:         "Okk",
		Completeness: effects.CompletenessFull,
		AttackLimits: []game.AttackLimit{effects.CantAttackUnlessACreatureWithGreaterPowerAttacks()},
		BlockRules:   []game.BlockRule{effects.CantBlockUnlessACreatureWithGreaterPowerBlocks()},
	})
}
