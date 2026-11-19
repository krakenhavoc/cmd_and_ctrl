package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8307bfb0-3ab2-467f-b13c-ad4acadf319c",
		Name:         "Spikeshot Elder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}{R}: This creature deals damage equal to its power to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    ManaCost("{1}{R}{R}"),
			Targets: TargetAny(),
			Effect:  sourceDealsDamageEqualToItsPowerToTarget,
		}},
	})
}
