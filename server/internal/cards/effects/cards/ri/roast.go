package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roast — Sorcery {1}{R} (Dragons of Tarkir,
// #5551):
//
//	"Roast deals 5 damage to target creature without flying."
//
// Charring Bite's body (Twinmaw Stormbrood).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5e60a85-2f31-47ec-a95b-2e69b2d2c00e",
		Name:         "Roast",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets:      effects.TargetCreature("target creature without flying", effects.WithoutKeyword("flying")),
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}
