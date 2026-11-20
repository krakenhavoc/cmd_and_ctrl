package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bold Impaler — Creature — Vampire Knight {R}, 1/2 (Eldritch Moon, #5541):
//
//	"{2}{R}: This creature gets +2/+0 until end of turn."
//
// Hematite Golem's pump at {2}{R}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "374e6c1b-5a9a-4d3f-8bff-2f6bc0419391",
		Name:         "Bold Impaler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{R}"),
			Effect:  thisGetsUntilEndOfTurn(2, 0, "Bold Impaler — +2/+0 until end of turn"),
		}},
	})
}
