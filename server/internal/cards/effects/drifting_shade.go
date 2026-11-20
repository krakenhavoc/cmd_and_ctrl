package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drifting Shade — Creature — Shade {3}{B}, 1/1 (Magic 2012, #5573):
//
//	"Flying
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Dungeon Shade's text: flying, and a {B} pump on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fa3883c7-4fb4-48b9-847b-9a94d1988dd7",
		Name:            "Drifting Shade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    ManaCost("{B}"),
			Effect:  thisGetsUntilEndOfTurn(1, 1, "Drifting Shade — +1/+1 until end of turn"),
		}},
	})
}
