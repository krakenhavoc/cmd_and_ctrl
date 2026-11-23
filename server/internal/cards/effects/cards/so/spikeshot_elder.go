package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spikeshot Elder — Creature — Goblin Shaman {R}, 1/1 (Scars of Mirrodin,
// #5577):
//
//	"{1}{R}{R}: This creature deals damage equal to its power to any
//	 target."
//
// Burning Anger's body (sourceDealsDamageEqualToItsPowerToTarget): the
// power is read as the ability resolves, or as the Elder last existed on
// the battlefield if it has left (CR 608.2h). No tap, so it can be
// activated again and again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8307bfb0-3ab2-467f-b13c-ad4acadf319c",
		Name:         "Spikeshot Elder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}: This creature deals damage equal to its power to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{1}{R}{R}"),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsDamageEqualToItsPowerToTarget,
		}},
	})
}
