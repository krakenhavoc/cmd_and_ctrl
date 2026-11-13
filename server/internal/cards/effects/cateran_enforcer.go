package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cateran Enforcer — Creature — Horror Mercenary {3}{B}{B}, 4/3
// (Mercadian Masques, #5619):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 {4}, {T}: Search your library for a Mercenary permanent card with
//	 mana value 4 or less, put it onto the battlefield, then shuffle."
//
// The Ramosian Rebels' search with the word Mercenary, at 4
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Enforcer waits out summoning sickness (CR
// 302.6). Fear is the printed keyword the block check reads (CR 702.36).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a2f3139e-80f6-43fb-bdb7-d45d6eb2a652",
		Name:            "Cateran Enforcer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Search your library for a Mercenary permanent card with mana value 4 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect: searchForASubtypePermanentWithManaValueAtMost("Mercenary", 4,
				"Cateran Enforcer — a Mercenary permanent card with mana value 4 or less, onto the battlefield"),
		}},
	})
}
