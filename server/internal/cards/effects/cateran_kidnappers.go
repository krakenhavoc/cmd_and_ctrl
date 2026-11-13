package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cateran Kidnappers — Creature — Human Mercenary {2}{B}{B}, 4/2
// (Mercadian Masques, #5619):
//
//	"{3}, {T}: Search your library for a Mercenary permanent card with
//	 mana value 3 or less, put it onto the battlefield, then shuffle."
//
// The Ramosian Rebels' search with the word Mercenary, at 3
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Kidnappers wait out summoning sickness (CR
// 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "938c4f2e-6dda-4ac4-b50c-40dc165b4eaa",
		Name:         "Cateran Kidnappers",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Search your library for a Mercenary permanent card with mana value 3 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Effect: searchForASubtypePermanentWithManaValueAtMost("Mercenary", 3,
				"Cateran Kidnappers — a Mercenary permanent card with mana value 3 or less, onto the battlefield"),
		}},
	})
}
