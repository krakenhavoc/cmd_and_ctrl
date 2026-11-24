package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hexdrinker — Creature — Snake {G}, 2/1 (Level up {1}):
//
//	"LEVEL 3-7: 4/4, protection from instants.
//	 LEVEL 8+: 6/6, protection from everything."
//
// A leveler (CR 702.87, CR 711): level up is the sorcery-speed activation and
// each band is the permanent's own static, so a Hexdrinker with no abilities is
// its printed 2/1 however many level counters it has. The protection keywords
// are the engine's own (the same strings Emrakul, the Promised End and
// Progenitus print).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69bc2afd-9f53-47f2-b9c8-f12732784e10",
		Name:         "Hexdrinker",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{1}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 3, To: 7, Power: 4, Toughness: 4, Keywords: []string{"protection from instants"}},
			effects.LevelBand{From: 8, Power: 6, Toughness: 6, Keywords: []string{"protection from everything"}},
		),
	})
}
