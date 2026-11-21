package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Direct Hit — Sorcery {2}{R} (Star Trek Commander, #4974):
//
//	"Direct Hit deals 5 damage to target creature."
//
// Concentrated Fire's body at sorcery speed (the timing is the card
// type's). A target that is gone or no longer a creature as it resolves
// takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6b0adcb-4676-4d9f-83ae-a2ff3cdfd84b",
		Name:         "Direct Hit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}
