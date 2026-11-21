package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rouse the Mob — Instant {R} (Journey into Nyx, #5558):
//
//	"Strive — This spell costs {2}{R} more to cast for each target
//	 beyond the first.
//	 Any number of target creatures each get +2/+0 and gain trample
//	 until end of turn."
//
// Strive is Call the Coppercoats' coloured per-target surcharge (CR
// 207.2c). Zero targets is legal; each target still legal as the spell
// resolves gets the effect, and one that left does not (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc8b2b89-7118-4662-beda-9108cd08491b",
		Name:         "Rouse the Mob",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "trample")),
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{2}{R}", "Strive — This spell costs {2}{R} more to cast for each target beyond the first."),
		},
		OnResolve: effects.JouStrivePump(2, 0, "Rouse the Mob — +2/+0 and trample", "trample"),
	})
}
