package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furnace Whelp — Creature — Dragon {2}{R}{R}, 2/2 (Fifth Dawn, #5604):
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Shivan Dragon's {R} firebreathing on a smaller flier, with the shared
// thisGetsUntilEndOfTurn body. No tap in the cost, so it can be activated
// any number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8422f1e0-00ca-4ffb-a6b2-e2c9f96d7f23",
		Name:            "Furnace Whelp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Furnace Whelp — +1/+0 until end of turn"),
		}},
	})
}
