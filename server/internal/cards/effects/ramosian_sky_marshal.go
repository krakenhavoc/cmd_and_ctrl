package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ramosian Sky Marshal — Creature — Human Rebel {3}{W}{W}, 3/3
// (Mercadian Masques, #5619):
//
//	"Flying
//	 {7}, {T}: Search your library for a Rebel permanent card with mana
//	 value 6 or less, put it onto the battlefield, then shuffle."
//
// Ramosian Sergeant's rebel search at 6
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Sky Marshal waits out summoning sickness (CR
// 302.6). Flying is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "52fcc5b6-6ada-457f-be76-fa674c74b4fd",
		Name:            "Ramosian Sky Marshal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{7}, {T}: Search your library for a Rebel permanent card with mana value 6 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{7}"), TapCost()),
			Effect: searchForASubtypePermanentWithManaValueAtMost("Rebel", 6,
				"Ramosian Sky Marshal — a Rebel permanent card with mana value 6 or less, onto the battlefield"),
		}},
	})
}
