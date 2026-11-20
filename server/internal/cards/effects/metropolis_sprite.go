package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metropolis Sprite — Creature — Faerie Rogue {1}{U}, 1/2 (Gatecrash, #5565):
//
//	"Flying
//	 {U}: This creature gets +1/-1 until end of turn."
//
// Breathstealer's pump at a blue cost; flying is the engine keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "924c7c4e-9c4b-496a-ba49-7fb3918a8d07",
		Name:            "Metropolis Sprite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{U}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{U}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Metropolis Sprite — +1/-1 until end of turn"),
		}},
	})
}
