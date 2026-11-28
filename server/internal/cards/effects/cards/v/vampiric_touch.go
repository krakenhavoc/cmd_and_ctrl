package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampiric Touch — Sorcery {2}{B} (Portal, S82 OLD batch 782, #5630):
//
//	"Vampiric Touch deals 2 damage to target opponent or planeswalker and you
//	 gain 2 life."
//
// Stolen Grain's shape with 2: the damage, then the life (CR 608.2c), nothing
// at all when the only target is illegal on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b701378f-681b-4840-bd11-1a97c2a675f2",
		Name:         "Vampiric Touch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetOpponentOrPlaneswalker(),
		Purpose:      game.Purpose{LifeGain: 2, Targets: effects.ForTargets(effects.DamageToTarget(0, 2)).Targets},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 2), effects.GainLife{Amount: 2}),
	})
}
