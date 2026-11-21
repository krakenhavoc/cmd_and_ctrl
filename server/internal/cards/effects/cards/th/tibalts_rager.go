package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tibalt's Rager — Creature — Devil {1}{R}, 1/2 (War of the Spark, #5520):
//
//	"When this creature dies, it deals 1 damage to any target.
//	 {1}{R}: This creature gets +2/+0 until end of turn."
//
// Munitions Enthusiast's dies trigger and Hematite Golem's pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d7d8f4f-abf3-408b-a289-1dd6a3ebdbf6",
		Name:         "Tibalt's Rager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Tibalt's Rager — 1 damage to any target", effects.SourceDealsDamageToEachLegalTarget(1)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Tibalt's Rager — +2/+0 until end of turn"),
		}},
	})
}
