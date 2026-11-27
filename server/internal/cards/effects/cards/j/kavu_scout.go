package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kavu Scout — Creature — Kavu Scout {2}{R}, 0/2 (Invasion, #5616):
//
//	"Domain — This creature gets +1/+0 for each basic land type among lands
//	 you control."
//
// A layer 7c static (S82OldB660SelfPumpDomain), its count read on every pass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2bd91ce8-1797-4dd7-8150-0d4caa7dad61",
		Name:         "Kavu Scout",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB660SelfPumpDomain(1, 0, "Domain — This creature gets +1/+0 for each basic land type among lands you control."),
		},
	})
}
