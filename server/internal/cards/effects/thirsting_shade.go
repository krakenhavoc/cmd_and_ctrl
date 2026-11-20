package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thirsting Shade — Creature — Shade {B}, 1/1 (Ravnica Allegiance, #5521):
//
//	"Lifelink
//	 {2}{B}: This creature gets +1/+1 until end of turn."
//
// Lifelink is printed data (CR 702.15b); the pump is Masked Blackguard's
// {2}{B} +1/+1 until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cba60174-a4de-41e1-a4d9-8f7bb1387e0e",
		Name:            "Thirsting Shade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    ManaCost("{2}{B}"),
			Effect:  thisGetsUntilEndOfTurn(1, 1, "Thirsting Shade — +1/+1 until end of turn"),
		}},
	})
}
