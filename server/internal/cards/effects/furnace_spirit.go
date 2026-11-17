package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furnace Spirit — Creature — Spirit {2}{R}, 1/1 (Stronghold, #5627):
//
//	"Haste
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Haste is a printed keyword. The pump is Firefly's +1/+0 at {R}, with
// the shared thisGetsUntilEndOfTurn body; each boost ends in the cleanup
// step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86b33d25-6123-439c-9026-a4e835b48562",
		Name:            "Furnace Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Furnace Spirit — +1/+0 until end of turn"),
		}},
	})
}
