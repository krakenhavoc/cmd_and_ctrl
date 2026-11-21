package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ornery Kudu — Creature — Antelope {2}{G}, 3/4
// (Amonkhet, #5535):
//
//	"When this creature enters, put a -1/-1 counter on target creature you control."
//
// Channeler Initiate's enters trigger with one counter; the target may
// be this creature itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8abef87a-4a5c-412e-8137-827cee50bb8b",
		Name:         "Ornery Kudu",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Ornery Kudu — put a -1/-1 counter on target creature you control", effects.Cmdr2B25CountersOnTarget(1)),
				effects.TargetCreature("target creature you control", effects.YouControl()),
			),
		},
	})
}
