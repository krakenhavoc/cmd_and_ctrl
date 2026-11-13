package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cateran Brute — Creature — Horror Mercenary {2}{B}, 2/2
// (Mercadian Masques, #5619):
//
//	"{2}, {T}: Search your library for a Mercenary permanent card with
//	 mana value 2 or less, put it onto the battlefield, then shuffle."
//
// The Ramosian Rebels' search with the word Mercenary, at 2
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Brute waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f60d03ae-f45f-45b2-a0f7-108acb0a45db",
		Name:         "Cateran Brute",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Search your library for a Mercenary permanent card with mana value 2 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: searchForASubtypePermanentWithManaValueAtMost("Mercenary", 2,
				"Cateran Brute — a Mercenary permanent card with mana value 2 or less, onto the battlefield"),
		}},
	})
}
