package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enslaved Scout — Creature — Goblin Scout {2}{R}, 2/2 (Alliances,
// #5634):
//
//	"{2}: This creature gains mountainwalk until end of turn. (It can't
//	 be blocked as long as defending player controls a Mountain.)"
//
// A layer 6 grant (CR 613.1f) pinned to the Scout that ends at cleanup
// (CR 514.2). Mountainwalk is landwalk, which the block check reads off
// the defending player's lands (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f39a3798-1ff0-4d1b-933f-bfcc0f265168",
		Name:         "Enslaved Scout",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gains mountainwalk until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("mountainwalk", "Enslaved Scout — mountainwalk until end of turn"),
		}},
	})
}
