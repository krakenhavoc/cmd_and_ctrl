package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veiled Shade — Creature — Shade {2}{B}, 2/2 (Guilds of Ravnica, #5525):
//
//	"{1}{B}: This creature gets +1/+1 until end of turn."
//
// Battlefield Percher's pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e9189d9-3b25-476b-b096-bb540398c1d1",
		Name:         "Veiled Shade",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    ManaCost("{1}{B}"),
			Effect:  thisGetsUntilEndOfTurn(1, 1, "Veiled Shade — +1/+1 until end of turn"),
		}},
	})
}
