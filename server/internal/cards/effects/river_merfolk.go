package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// River Merfolk — Creature — Merfolk {U}{U}, 2/1 (Fallen Empires,
// #5637):
//
//	"{U}: This creature gains mountainwalk until end of turn. (It can't
//	 be blocked as long as defending player controls a Mountain.)"
//
// A layer-6 grant of mountainwalk on the creature itself until end of
// turn (CR 702.14), Zombie Trailblazer's swampwalk on its own source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f503ee5-9180-438e-8af5-2af7f8f08a3a",
		Name:         "River Merfolk",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}: This creature gains mountainwalk until end of turn. (It can't be blocked as long as defending player controls a Mountain.)",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{U}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("mountainwalk", "River Merfolk — mountainwalk until end of turn"),
		}},
	})
}
