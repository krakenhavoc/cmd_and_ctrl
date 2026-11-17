package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cateran Persuader — Creature — Human Mercenary {B}{B}, 2/1
// (Mercadian Masques, #5619):
//
//	"{1}, {T}: Search your library for a Mercenary permanent card with
//	 mana value 1 or less, put it onto the battlefield, then shuffle."
//
// The Ramosian Rebels' search with the word Mercenary, at 1
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Persuader waits out summoning sickness (CR
// 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "415e9484-0788-4c00-bb1f-19308511114b",
		Name:         "Cateran Persuader",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Search your library for a Mercenary permanent card with mana value 1 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect: searchForASubtypePermanentWithManaValueAtMost("Mercenary", 1,
				"Cateran Persuader — a Mercenary permanent card with mana value 1 or less, onto the battlefield"),
		}},
	})
}
