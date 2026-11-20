package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horror of the Dim — Creature — Horror {4}{B}, 3/4 (Gatecrash, #5565):
//
//	"{U}: This creature gains hexproof until end of turn. (It can't be the
//	 target of spells or abilities your opponents control.)"
//
// A layer-6 hexproof grant pinned to this creature until cleanup
// (thisGainsKeywordUntilEndOfTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83e91e0a-2680-4bd3-98eb-7567791e77d5",
		Name:         "Horror of the Dim",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}: This creature gains hexproof until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{U}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("hexproof", "Horror of the Dim — hexproof until end of turn"),
		}},
	})
}
