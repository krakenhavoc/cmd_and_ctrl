package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pavel Maliki — Legendary Creature — Human {4}{B}{R}, 5/3 (Legends, #5639):
//
//	"{B}{R}: Pavel Maliki gets +1/+0 until end of turn."
//
// The shared legSelfPump row: no tap in the cost, so it can be activated
// any number of times, summoning sick or not, and the boost ends in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41026872-3c7b-40c7-8d0b-9d2f9b6c3e91",
		Name:         "Pavel Maliki",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LegSelfPump("{B}{R}", 0, "Pavel Maliki", "Pavel Maliki")},
	})
}
