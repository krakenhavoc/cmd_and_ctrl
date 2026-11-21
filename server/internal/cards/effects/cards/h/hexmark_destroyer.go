package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hexmark Destroyer — Artifact Creature — Necron {4}{B}{B}, 6/6 (Warhammer
// 40,000 Commander, S82 GAP batch 187, #5948):
//
//	"Multi-threat Eliminator — This creature can't be blocked except by six
//	 or more creatures.
//	 Unearth {4}{B}{B}"
//
// Pathrazer of Ulamog's block rule at six.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3329240-c812-45ea-a7f8-4d79cd4ed515",
		Name:         "Hexmark Destroyer",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.MinBlockers(effects.OnSelf(), 6)},
		Activated:    []effects.ActivatedAbility{effects.Unearth("{4}{B}{B}")},
	})
}
