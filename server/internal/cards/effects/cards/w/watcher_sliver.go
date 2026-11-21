package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Watcher Sliver — Creature — Sliver {3}{W}, 2/2 (Time Spiral, #5596):
//
//	"All Sliver creatures get +0/+2."
//
// The defensive anthem: every Sliver creature at the table gets +0/+2,
// this one and an opponent's included, in layer 7c (CR 613.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a538bd7-e665-4896-95df-bbab1e28f75e",
		Name:         "Watcher Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Sliver"}}, 0, 2),
		},
	})
}
