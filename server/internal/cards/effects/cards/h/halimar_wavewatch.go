package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halimar Wavewatch — {1}{U} Creature — Merfolk Soldier 0/3 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Level up {2}
//	LEVEL 1-4: 0/6
//	LEVEL 5+: 6/6, Islandwalk
//
// A leveler: Level up is the keyword, each level symbol a band of LevelBands.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "44f01491-cce4-442e-a8d9-345a074805bd",
		Name:         "Halimar Wavewatch",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{2}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 4, Power: 0, Toughness: 6},
			effects.LevelBand{From: 5, Power: 6, Toughness: 6, Keywords: []string{"islandwalk"}},
		),
	})
}
