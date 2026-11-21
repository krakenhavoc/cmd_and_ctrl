package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Engulfing Eruption — Sorcery {2}{R}{R} (Core Set 2020, #5518):
//
//	"Engulfing Eruption deals 5 damage to target creature."
//
// Direct Hit's body at sorcery speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27b4fee0-ae46-4d15-96cd-aff8f3009de8",
		Name:         "Engulfing Eruption",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}
