package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampiric Feast — Sorcery {5}{B}{B} (Portal, #5630):
//
//	"Vampiric Feast deals 4 damage to any target and you gain 4 life."
//
// Void Helix's body with 4: the damage, then the life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1980ca2e-a415-4de1-ac30-7055507e82a2",
		Name:         "Vampiric Feast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.S82B48DamageToTheFirstTarget(4)),
			effects.GainLife{Amount: 4},
		),
	})
}
