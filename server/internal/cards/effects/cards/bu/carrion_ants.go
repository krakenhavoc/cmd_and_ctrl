package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carrion Ants — Creature — Insect {2}{B}{B}, 0/1 (Legends, #5639):
//
//	"{1}: This creature gets +1/+1 until end of turn."
//
// The shared legSelfPump row: no tap in the cost, so it can be activated
// any number of times, summoning sick or not, and the boost ends in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87bc6be7-4734-4a11-806e-e9e4f5616fe0",
		Name:         "Carrion Ants",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LegSelfPump("{1}", 1, "This creature", "Carrion Ants")},
	})
}
