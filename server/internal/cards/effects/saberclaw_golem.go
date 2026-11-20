package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saberclaw Golem — Artifact Creature — Golem {5}, 4/2 (Scars of Mirrodin, #5577):
//
//	"{R}: This creature gains first strike until end of turn."
//
// Noble Panther's ability at {R}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4989d38e-332f-48f5-863b-1566921156c8",
		Name:         "Saberclaw Golem",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    ManaCost("{R}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("first strike", "Saberclaw Golem — first strike until end of turn"),
		}},
	})
}
