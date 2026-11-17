package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ramosian Captain — Creature — Human Rebel {1}{W}{W}, 2/2
// (Mercadian Masques, #5619):
//
//	"First strike
//	 {5}, {T}: Search your library for a Rebel permanent card with mana
//	 value 4 or less, put it onto the battlefield, then shuffle."
//
// Ramosian Sergeant's rebel search at 4
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Captain waits out summoning sickness (CR
// 302.6). First strike is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "30255f36-2e26-4990-8b54-27f38130d45b",
		Name:            "Ramosian Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Search your library for a Rebel permanent card with mana value 4 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			Effect: searchForASubtypePermanentWithManaValueAtMost("Rebel", 4,
				"Ramosian Captain — a Rebel permanent card with mana value 4 or less, onto the battlefield"),
		}},
	})
}
