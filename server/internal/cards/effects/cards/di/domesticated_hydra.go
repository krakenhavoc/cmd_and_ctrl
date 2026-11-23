package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Domesticated Hydra — 3/3 Hydra for {2}{G}{G}:
//
//	"{X}{G}{G}{G}: Monstrosity X. (If this creature isn't monstrous,
//	 put X +1/+1 counters on it and it becomes monstrous.)
//	 As long as this creature is monstrous, it has trample."
//
// Monstrosity X with the announced X, and a gated trample (#1700).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6d65fa3-26a9-44b2-a93d-a295bd080729",
		Name:         "Domesticated Hydra",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.MonstrosityX(effects.ManaCost("{X}{G}{G}{G}"))},
		Static:       []game.StaticAbility{effects.MonstrousKeywords("trample")},
	})
}
