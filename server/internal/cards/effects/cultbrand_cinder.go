package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cbb030aa-ba4f-4d9e-920d-001a40982f5e",
		Name:         "Cultbrand Cinder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Cultbrand Cinder — put a -1/-1 counter on target creature", cmdr2B25CountersOnTarget(1)),
				Harmful(TargetCreature("target creature")),
			),
		},
	})
}
