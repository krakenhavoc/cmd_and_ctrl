package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreadwing — Creature — Zombie {B}, 1/1 (Conflux, #5587):
//
//	"{1}{U}{R}: This creature gets +3/+0 and gains flying until end of
//	 turn."
//
// thisCreatureUntilEOT's pump and grant on the source, both ending in
// the cleanup step (CR 514.2). With no {T} it can be activated any
// number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "446b9626-2e64-4933-afc8-891135b8e3d6",
		Name:         "Dreadwing",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}{U}{R}: This creature gets +3/+0 and gains flying until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump | game.AnswerCombatGrant,
				Pump:    &game.Pump{Power: 3, Keywords: []string{"flying"}},
			},
			Cost:   ManaCost("{1}{U}{R}"),
			Effect: thisCreatureUntilEOT("Dreadwing — +3/+0 and flying until end of turn", 3, 0, "flying"),
		}},
	})
}
