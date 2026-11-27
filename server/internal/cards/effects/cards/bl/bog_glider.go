package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Glider — Creature — Human Mercenary {2}{B}, 1/1 (Prophecy, S82 OLD batch
// 675, #5617):
//
//	"Flying
//	 {T}, Sacrifice a land: Search your library for a Mercenary permanent card
//	 with mana value 2 or less, put it onto the battlefield, then shuffle."
//
// Flying is a printed keyword. The ability is the Nemesis recruiters' search
// (SearchForASubtypePermanentWithManaValueAtMost, as Defiant Falcon's), paid
// with {T} and a land instead of mana: the searcher may find nothing
// (CR 701.23b) and the library is shuffled either way. The {T} is a
// creature's, so it waits out summoning sickness (CR 302.6). The land is the
// cost (Spitting Spider's B08SacrificeALand), so it is gone even if the ability
// is countered.
//
// No simplification.
func init() {
	const what = "a Mercenary permanent card with mana value 2 or less"
	effects.Register(effects.Spec{
		OracleID:        "6bff9d3a-a98d-4060-8409-4b81ce5aab15",
		Name:            "Bog Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a land: Search your library for " + what + ", put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.B08SacrificeALand()),
			Effect:  effects.SearchForASubtypePermanentWithManaValueAtMost("Mercenary", 2, "Bog Glider — "+what+", onto the battlefield"),
		}},
	})
}
