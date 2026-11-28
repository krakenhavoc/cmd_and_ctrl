package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Cannon — Artifact {2} (Tempest, S82 OLD b755, #5628):
//
//	"{T}: Target creature you control gets +1/+0 and gains flying until end
//	 of turn. Destroy that creature at the beginning of the next end step."
//
// Stone Giant's grant-then-destroy shape (OldB755TargetGetsThenDestroyAtEndStep):
// the delayed trigger destroys the creature as it is then, if it is still the
// same object (CR 603.7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "710a51e1-a305-49b5-a429-0ff1a10244fa",
		Name:         "Mogg Cannon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature you control gets +1/+0 and gains flying until end of turn. Destroy that creature at the beginning of the next end step.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "flying")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  effects.OldB755TargetGetsThenDestroyAtEndStep("Mogg Cannon", 1, "flying"),
		}},
	})
}
