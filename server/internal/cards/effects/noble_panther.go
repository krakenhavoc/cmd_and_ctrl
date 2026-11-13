package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Noble Panther — Creature — Cat {1}{G}{W}, 3/3 (Invasion, #5616):
//
//	"{1}: This creature gains first strike until end of turn."
//
// Zhalfirin Knight's row at {1}: a layer-6 grant pinned to the Panther
// itself. There is no {T} in the cost, so it can be activated the turn
// the Panther arrives, and more than once (a second activation changes
// nothing).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bfb799d2-b2fa-4efb-91d7-a457230b370f",
		Name:         "Noble Panther",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    ManaCost("{1}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("first strike", "Noble Panther — first strike until end of turn"),
		}},
	})
}
