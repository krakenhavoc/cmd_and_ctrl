package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marvelous Melee — Instant {4}{R} (Marvel Super Heroes Commander,
// #4455): "Marvelous Melee deals 6 damage to target creature. Basic
// landcycling {2}". Concentrated Fire's damage body at six damage, with
// the cycling family's shared BasicLandcycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ee9940a-0569-45b6-b28c-0ef3914013a9",
		Name:         "Marvelous Melee",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 6)),
		OnResolve:    effects.DamageToFirstTargetStep(6),
		Activated:    []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
