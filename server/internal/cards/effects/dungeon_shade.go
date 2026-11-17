package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dungeon Shade — Creature — Shade Spirit {3}{B}, 1/1 (Stronghold,
// #5627):
//
//	"Flying
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Flying is a printed keyword. The pump is the Shade shape at {B}, with
// the shared thisGetsUntilEndOfTurn body: it can be activated any number
// of times, summoning sick or not, and each boost ends in the cleanup
// step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "953a1a07-1bf2-480e-81a3-f2d5f19f6df5",
		Name:            "Dungeon Shade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    ManaCost("{B}"),
			Effect:  thisGetsUntilEndOfTurn(1, 1, "Dungeon Shade — +1/+1 until end of turn"),
		}},
	})
}
