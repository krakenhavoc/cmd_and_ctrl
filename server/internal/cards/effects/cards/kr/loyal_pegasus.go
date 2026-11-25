package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loyal Pegasus — Creature — Pegasus {W}, 2/1 (Core Set 2020, #5518):
//
//	"Flying
//	 This creature can't attack or block alone."
//
// Mogg Flunkies's two restrictions on the creature itself (CR 506.5, 508.1c,
// 509.1b), with a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3b7b862a-7edd-4e96-b4b9-f7113b4c93f9",
		Name:            "Loyal Pegasus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AttackLimits:    []game.AttackLimit{effects.CantAttackAlone()},
		BlockRules:      []game.BlockRule{effects.CantBlockAlone()},
	})
}
