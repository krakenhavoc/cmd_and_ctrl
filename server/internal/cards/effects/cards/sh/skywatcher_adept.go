package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skywatcher Adept — {U} Creature — Merfolk Wizard 1/1 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Level up {3}
//	LEVEL 1-2: 2/2, Flying
//	LEVEL 3+: 4/2, Flying
//
// A leveler: Level up is the keyword, each level symbol a band of LevelBands.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d63725af-be6d-4f91-b5c8-04df82d6f805",
		Name:         "Skywatcher Adept",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{3}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 2, Power: 2, Toughness: 2, Keywords: []string{"flying"}},
			effects.LevelBand{From: 3, Power: 4, Toughness: 2, Keywords: []string{"flying"}},
		),
	})
}
