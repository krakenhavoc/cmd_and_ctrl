package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ikiral Outrider — {1}{W} Creature — Human Soldier 1/2 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Level up {4}
//	LEVEL 1-3: 2/6, Vigilance
//	LEVEL 4+: 3/10, Vigilance
//
// A leveler: Level up is the keyword, each level symbol a band of LevelBands.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c03e0db-ce26-41f6-a12c-861a919e52fb",
		Name:         "Ikiral Outrider",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{4}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 3, Power: 2, Toughness: 6, Keywords: []string{"vigilance"}},
			effects.LevelBand{From: 4, Power: 3, Toughness: 10, Keywords: []string{"vigilance"}},
		),
	})
}
