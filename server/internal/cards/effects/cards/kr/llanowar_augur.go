package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Augur — Creature — Elf Shaman {G}, 0/3 (Future Sight, #5594):
//
//	"Sacrifice this creature: Target creature gets +3/+3 and gains
//	 trample until end of turn. Activate only during your upkeep."
//
// An upkeep-only sacrifice (DuringYourUpkeep is the step AND your turn)
// with the shared pump-and-grant body on its one target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3052531-f7cf-44a7-9492-eb155d75268c",
		Name:         "Llanowar Augur",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "Sacrifice this creature: Target creature gets +3/+3 and gains trample until end of turn. Activate only during your upkeep.",
			Purpose:   effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "trample")),
			Cost:      effects.SacrificeThis(),
			Condition: effects.DuringYourUpkeep(),
			Targets:   effects.TargetCreature("target creature"),
			Effect:    effects.PumpAndGrantFirstTarget(3, 3, "Llanowar Augur — +3/+3 and trample until end of turn", "trample"),
		}},
	})
}
