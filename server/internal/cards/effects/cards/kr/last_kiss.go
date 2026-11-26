package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Last Kiss — Instant {2}{B} (Rise of the Eldrazi, S82 OLD batch 483, #5581):
//
//	"Last Kiss deals 2 damage to target creature and you gain 2 life."
//
// Sorin's Thirst's body at another price: the damage, then the life gain
// (which happens even if the creature has gone, CR 608.2b only skipping the
// damage if the one target is illegal, and then the spell is countered).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc59bf58-373d-4562-9a67-0721fcf2f098",
		Name:         "Last Kiss",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), LifeGain: 2, Answers: game.AnswerRemove},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 2), effects.GainLife{Amount: 2}),
	})
}
