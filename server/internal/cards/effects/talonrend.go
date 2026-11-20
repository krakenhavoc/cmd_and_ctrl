package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Talonrend — Creature — Elemental {4}{U}, 0/5 (Eventide, #5590):
//
//	"Flying
//	 {U/R}: This creature gets +1/-1 until end of turn."
//
// Frostburn Weird's pump on a flyer; the hybrid symbol is paid with
// either colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "85c6290e-2bf2-4666-a606-326a8209ace6",
		Name:            "Talonrend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{U/R}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{U/R}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Talonrend — +1/-1 until end of turn"),
		}},
	})
}
