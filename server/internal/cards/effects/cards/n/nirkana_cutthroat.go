package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nirkana Cutthroat — {2}{B} Creature — Vampire Warrior 3/2 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Level up {2}{B}
//	LEVEL 1-2: 4/3, Deathtouch
//	LEVEL 3+: 5/4, First strike, deathtouch
//
// A leveler: Level up is the keyword, each level symbol a band of LevelBands.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "759774d2-abcf-4fc6-8433-5f251e80c70e",
		Name:         "Nirkana Cutthroat",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{2}{B}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 2, Power: 4, Toughness: 3, Keywords: []string{"deathtouch"}},
			effects.LevelBand{From: 3, Power: 5, Toughness: 4, Keywords: []string{"first strike", "deathtouch"}},
		),
	})
}
