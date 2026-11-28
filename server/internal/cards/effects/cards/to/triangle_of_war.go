package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triangle of War — Artifact {1} (Visions, S82 OLD batch 786, #5631):
//
//	"{2}, Sacrifice this artifact: Target creature you control fights target
//	 creature an opponent controls."
//
// Contested Cliffs' two-clause fight (SlotZeroFightsSlotOne): slot 0 is a
// creature you control, slot 1 a creature an opponent controls. The
// artifact is sacrificed as a cost, so it is gone before the targets are
// checked; if either target is illegal as the ability resolves, neither
// fights (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d4daf9c-e2f4-4d0e-b2e4-ded04d644b11",
		Name:         "Triangle of War",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, Sacrifice this artifact: Target creature you control fights target creature an opponent controls.",
			Cost:  effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			),
			Effect: effects.SlotZeroFightsSlotOne,
		}},
	})
}
