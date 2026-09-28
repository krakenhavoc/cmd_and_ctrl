package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b6d65fa3-26a9-44b2-a93d-a295bd080729",
		Name:         "Domesticated Hydra",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{MonstrosityX(ManaCost("{X}{G}{G}{G}"))},
		Static:       []game.StaticAbility{MonstrousKeywords("trample")},
	})
}
