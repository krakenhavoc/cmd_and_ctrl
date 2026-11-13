package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ramosian Lieutenant — Creature — Human Rebel {1}{W}, 1/2
// (Mercadian Masques, #5619):
//
//	"{4}, {T}: Search your library for a Rebel permanent card with mana
//	 value 3 or less, put it onto the battlefield, then shuffle."
//
// Ramosian Sergeant's rebel search at 3
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Lieutenant waits out summoning sickness (CR
// 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ddd2a87e-4020-4408-b4e6-a2120d485855",
		Name:         "Ramosian Lieutenant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Search your library for a Rebel permanent card with mana value 3 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect: searchForASubtypePermanentWithManaValueAtMost("Rebel", 3,
				"Ramosian Lieutenant — a Rebel permanent card with mana value 3 or less, onto the battlefield"),
		}},
	})
}
