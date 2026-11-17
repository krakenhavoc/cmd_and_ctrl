package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Capashen Knight — Creature — Human Knight {1}{W}, 1/1 (Urza's Destiny,
// #5621):
//
//	"First strike
//	 {1}{W}: This creature gets +1/+0 until end of turn."
//
// First strike rides PrintedKeywords and the combat damage step reads
// it. The pump is Avarax's at {1}{W}, through the shared
// thisGetsUntilEndOfTurn body, and it stacks with each activation until
// the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e295d207-128f-419a-866a-c9ac6248c199",
		Name:            "Capashen Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    ManaCost("{1}{W}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Capashen Knight — +1/+0 until end of turn"),
		}},
	})
}
