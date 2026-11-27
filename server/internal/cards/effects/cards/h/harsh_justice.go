package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harsh Justice — Instant {2}{W} (Portal, S82 OLD batch 782, #5630):
//
//	"Cast this spell only during the declare attackers step and only if you've
//	 been attacked this step.
//	 This turn, whenever an attacking creature deals combat damage to you, it
//	 deals that much damage to its controller."
//
// Champion's Victory's cast restriction. On resolution a repeating delayed
// trigger (CR 603.7b) is created for the rest of the turn, controlled by the
// caster ("you"): each combat damage event to the caster from an attacking
// creature puts one trigger on the stack, and the creature, as it last was
// (CR 608.2h), deals that much damage to its controller. The damage is dealt
// even if the creature has died by then, and by the creature as the source, so
// lifelink and protection read it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "fe4bee2c-f03f-44a4-94a4-55a06bcd0ad8",
		Name:               "Harsh Justice",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		OnResolve:          effects.S82OldB782HarshJusticeResolve,
	})
}
