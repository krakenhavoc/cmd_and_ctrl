package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cateran Slaver — Creature — Horror Mercenary {4}{B}{B}, 5/5
// (Mercadian Masques, #5619):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)
//	 {5}, {T}: Search your library for a Mercenary permanent card with
//	 mana value 5 or less, put it onto the battlefield, then shuffle."
//
// The Ramosian Rebels' search with the word Mercenary, at 5
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Slaver waits out summoning sickness (CR
// 302.6). Swampwalk is a printed keyword, enforced by the block check
// (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7989ef3-d2a5-41c6-944c-f969d850f557",
		Name:            "Cateran Slaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Search your library for a Mercenary permanent card with mana value 5 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			Effect: searchForASubtypePermanentWithManaValueAtMost("Mercenary", 5,
				"Cateran Slaver — a Mercenary permanent card with mana value 5 or less, onto the battlefield"),
		}},
	})
}
