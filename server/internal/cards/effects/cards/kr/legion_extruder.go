package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Legion Extruder — Artifact {1}{R} (The Big Score, #4468):
//
//	"When this artifact enters, it deals 2 damage to any target.
//	 {2}, {T}, Sacrifice another artifact: Create a 3/3 colorless Golem
//	 artifact creature token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c77dc11e-d95d-4e7e-bb24-d971c179e9b0",
		Name:         "Legion Extruder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Legion Extruder — 2 damage to any target",
				effects.SourceDealsDamageToEachLegalTarget(2)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice another artifact: Create a 3/3 colorless Golem artifact creature token.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another artifact", effects.Artifact())),
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Effect:  effects.Do(effects.CreateToken{Template: effects.B35GolemToken(), N: 1}),
		}},
	})
}
