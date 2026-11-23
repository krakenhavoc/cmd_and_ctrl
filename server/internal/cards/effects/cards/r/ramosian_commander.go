package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ramosian Commander — Creature — Human Rebel {2}{W}{W}, 2/4
// (Mercadian Masques, #5619):
//
//	"{6}, {T}: Search your library for a Rebel permanent card with mana
//	 value 5 or less, put it onto the battlefield, then shuffle."
//
// Ramosian Sergeant's rebel search at 5
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Commander waits out summoning sickness (CR
// 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05baf38c-677a-42e3-bb8c-a52a0a478984",
		Name:         "Ramosian Commander",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}, {T}: Search your library for a Rebel permanent card with mana value 5 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost()),
			Effect: effects.SearchForASubtypePermanentWithManaValueAtMost("Rebel", 5,
				"Ramosian Commander — a Rebel permanent card with mana value 5 or less, onto the battlefield"),
		}},
	})
}
