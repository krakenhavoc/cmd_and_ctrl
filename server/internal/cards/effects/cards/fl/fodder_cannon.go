package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fodder Cannon — Artifact {4} (Urza's Destiny, #5621):
//
//	"{4}, {T}, Sacrifice a creature: This artifact deals 4 damage to
//	 target creature."
//
// Disruptor Pistol's shot, paid for with a creature instead of the
// artifact itself, so the Cannon stays and fires again once it untaps.
// The creature goes as a cost (CR 701.21a); a target gone by resolution
// takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaf171bd-a4bb-4ce4-836a-da193c94f42e",
		Name:         "Fodder Cannon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice a creature: This artifact deals 4 damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeACreature()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.WithRiders(effects.ForTargets(effects.DamageToTarget(0, 4)), game.Purpose{Answers: game.AnswerSacOutlet}),
			Effect:  effects.DealDamageToTheTarget(4),
		}},
	})
}
