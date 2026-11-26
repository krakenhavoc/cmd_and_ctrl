package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Cliffhaven — Creature — Kor Knight {1}{W}, 2/2 (Duel Decks: Zendikar vs. Eldrazi, #5548):
//
//	"Level up {3}
//	 LEVEL 1-3: 2/3, Flying
//	 LEVEL 4+: 4/4, Flying, vigilance"
//
// A leveler (CR 702.87, CR 711): the bands are statics gated on the level
// counters, and the printed size is the card's below the first band. Level
// up is an activated ability usable only as a sorcery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5045cfc3-ea63-4e0d-bf6a-bdeaf841ae3c",
		Name:         "Knight of Cliffhaven",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{3}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 3, Power: 2, Toughness: 3, Keywords: []string{"flying"}},
			effects.LevelBand{From: 4, Power: 4, Toughness: 4, Keywords: []string{"flying", "vigilance"}},
		),
	})
}
