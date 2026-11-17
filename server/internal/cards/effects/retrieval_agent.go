package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Retrieval Agent — Creature — Human Soldier {3}{U}, 2/5 (The Brothers' War, #5482):
//
//	"{2}: This creature gets +1/-1 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a6a1e58-85a9-436e-b1a0-373baa44d6f2",
		Name:         "Retrieval Agent",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{2}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Retrieval Agent — +1/-1 until end of turn"),
		}},
	})
}
