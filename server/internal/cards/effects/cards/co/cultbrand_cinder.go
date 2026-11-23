package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cultbrand Cinder — Creature — Elemental Shaman {4}{B/R}, 3/3
// (Shadowmoor, #5591):
//
//	"When this creature enters, put a -1/-1 counter on target creature."
//
// Dread Tiller's enters trigger: the target is chosen as the trigger goes
// on the stack (CR 603.3d) and re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbb030aa-ba4f-4d9e-920d-001a40982f5e",
		Name:         "Cultbrand Cinder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Cultbrand Cinder — put a -1/-1 counter on target creature", effects.Cmdr2B25CountersOnTarget(1)),
				effects.Harmful(effects.TargetCreature("target creature")),
			),
		},
	})
}
