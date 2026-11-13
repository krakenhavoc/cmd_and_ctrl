package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keeper of Kookus — Creature — Goblin {R}, 1/1 (Visions, #5631):
//
//	"{R}: This creature gains protection from red until end of turn."
//
// A layer 6 grant pinned to the Keeper (CR 613.1f) that ends at cleanup
// (CR 514.2). While it lasts, red spells and abilities from red sources
// can't target the Keeper, and damage from red sources is prevented
// (CR 702.16b, CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97417270-e525-4261-b7d9-0dde3360e666",
		Name:         "Keeper of Kookus",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gains protection from red until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{R}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("protection from red", "Keeper of Kookus — protection from red until end of turn"),
		}},
	})
}
