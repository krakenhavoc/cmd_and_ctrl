package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ramosian Sergeant — Creature — Human Rebel {W}, 1/1
// (Mercadian Masques, #5619):
//
//	"{3}, {T}: Search your library for a Rebel permanent card with mana
//	 value 2 or less, put it onto the battlefield, then shuffle."
//
// The rebel search, the shared
// searchForASubtypePermanentWithManaValueAtMost body at 2: the searcher
// may find nothing (CR 701.23b), and the library is shuffled either
// way. The {T} is a creature's, so the Sergeant waits out summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b0a14b9-977d-4e25-a519-eaddb0f94b6a",
		Name:         "Ramosian Sergeant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Search your library for a Rebel permanent card with mana value 2 or less, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: effects.SearchForASubtypePermanentWithManaValueAtMost("Rebel", 2,
				"Ramosian Sergeant — a Rebel permanent card with mana value 2 or less, onto the battlefield"),
		}},
	})
}
