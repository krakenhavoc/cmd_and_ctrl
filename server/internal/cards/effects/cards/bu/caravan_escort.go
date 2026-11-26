package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caravan Escort — Creature — Human Knight {W}, 1/1 (Duel Decks: Zendikar vs. Eldrazi, #5548):
//
//	"Level up {2}
//	 LEVEL 1-4: 2/2
//	 LEVEL 5+: 5/5, First strike"
//
// A leveler (CR 702.87, CR 711): the bands are statics gated on the level
// counters, and the printed size is the card's below the first band. Level
// up is an activated ability usable only as a sorcery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a9c221d-9bb4-4ec4-b5db-cb26db6065df",
		Name:         "Caravan Escort",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{2}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 4, Power: 2, Toughness: 2},
			effects.LevelBand{From: 5, Power: 5, Toughness: 5, Keywords: []string{"first strike"}},
		),
	})
}
