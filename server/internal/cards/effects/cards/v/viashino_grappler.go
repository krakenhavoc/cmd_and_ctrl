package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Grappler (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Lizard {2}{R}, 3/1:
//	"{G}: This creature gains trample until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74c5bfb0-1375-46e9-9a6c-1b9fada2913b",
		Name:         "Viashino Grappler",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.S82OldB510ThisGainsRow("{G}", "trample")},
	})
}
