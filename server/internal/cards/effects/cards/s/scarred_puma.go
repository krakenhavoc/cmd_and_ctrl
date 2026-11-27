package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scarred Puma — Creature — Cat {R}, 2/1 (Invasion, S82 older sets batch 664, #5616):
//
//	"This creature can't attack unless a black or green creature also
//	 attacks."
//
// Okk's companion limit (CR 508.1c) with the companion narrowed to a black or a
// green creature, any player's attacking beside it, read through the layers
// at declaration: game.AttackLimit.OthersColors, the small seam this card adds.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e28e1921-797b-4e6e-8058-3929af46eeb5",
		Name:         "Scarred Puma",
		Completeness: effects.CompletenessFull,
		AttackLimits: []game.AttackLimit{effects.OldB664CantAttackUnlessACreatureOfColorAlsoAttacks("B", "G")},
	})
}
