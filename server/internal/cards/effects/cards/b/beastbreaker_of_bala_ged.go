package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beastbreaker of Bala Ged — Creature — Human Warrior {1}{G}, 2/2 (Duel Decks: Zendikar vs. Eldrazi, #5548):
//
//	"Level up {2}{G}
//	 LEVEL 1-3: 4/4
//	 LEVEL 4+: 6/6, Trample"
//
// A leveler (CR 702.87, CR 711): the bands are statics gated on the level
// counters, and the printed size is the card's below the first band. Level
// up is an activated ability usable only as a sorcery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c77f436-81ab-4200-b76d-58b1f91fc3d0",
		Name:         "Beastbreaker of Bala Ged",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{2}{G}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 3, Power: 4, Toughness: 4},
			effects.LevelBand{From: 4, Power: 6, Toughness: 6, Keywords: []string{"trample"}},
		),
	})
}
