package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herbal Poultice — Artifact {0} (Lorwyn, #5593):
//
//	"{3}, Sacrifice this artifact: Regenerate target creature."
//
// The artifact goes as part of the cost (CR 701.21a), so the shield lands
// even if the Poultice is gone by resolution. It replaces the target's next
// destruction this turn (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a8e1f8e-3740-4ff3-b7fa-390ea278fb95",
		Name:         "Herbal Poultice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.LrwB1RegenerateTargetRow("{3}, Sacrifice this artifact: Regenerate target creature.",
				effects.Plus(effects.ManaCost("{3}"), effects.SacrificeThis()), effects.TargetCreature("target creature")),
		},
	})
}
