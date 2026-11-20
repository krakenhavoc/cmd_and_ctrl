package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marble Gargoyle — Artifact Creature — Gargoyle {2}{W}, 2/2 (Modern
// Horizons 2, #5501):
//
//	"Flying
//	 {W}: This creature gets +0/+1 until end of turn."
//
// Honor Guard's +0/+1 at {W} (the shared thisGetsUntilEndOfTurn body,
// ending in the cleanup step, CR 514.2) on a flyer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8a1fd464-760c-43d9-829a-424945d91eb9",
		Name:            "Marble Gargoyle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{W}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    ManaCost("{W}"),
			Effect:  thisGetsUntilEndOfTurn(0, 1, "Marble Gargoyle — +0/+1 until end of turn"),
		}},
	})
}
