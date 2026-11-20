package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deranged Outcast — Creature — Human Rogue {1}{G}, 2/1 (Dark Ascension, #5571):
//
//	"{1}{G}, Sacrifice a Human: Put two +1/+1 counters on target
//	 creature."
//
// The Human is sacrificed as part of the cost (CR 602.2b); the Outcast is
// a Human itself and may pay with its own body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "115f1397-fa24-4412-80b7-e2a6ddedcd4c",
		Name:         "Deranged Outcast",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice a Human: Put two +1/+1 counters on target creature.",
			Cost:    game.AbilityCost{Mana: "{1}{G}", SacrificeOther: sacrificeSpec("a Human", HasSubtype("Human"))},
			Targets: TargetCreature("target creature"),
			Effect:  twoPlusOneCountersOnFirstLegalTarget,
		}},
	})
}
