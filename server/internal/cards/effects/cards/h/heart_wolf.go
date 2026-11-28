package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heart Wolf — Creature — Wolf {3}{R}, 2/2 (Homelands, S82 OLD b823, #5635):
//
//	"First strike
//	{T}: Target Dwarf creature gets +2/+0 and gains first strike until end of
//	turn. When that creature leaves the battlefield this turn, sacrifice this
//	creature. Activate only during combat."
//
// Runesword's shape: the pump and the grant run on the Dwarf, then a delayed
// trigger (CR 603.7) pinned to that Dwarf for the turn sacrifices the Wolf if it
// is still the same object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f50c45f-058d-41fb-9965-08c2e0edc8ac",
		Name:            "Heart Wolf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Target Dwarf creature gets +2/+0 and gains first strike until end of turn. When that creature leaves the battlefield this turn, sacrifice this creature. Activate only during combat.",
			Purpose:   effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "first strike")),
			Cost:      effects.TapCost(),
			Targets:   effects.TargetCreature("target Dwarf creature", effects.HasSubtype("Dwarf")),
			Condition: effects.OldB823DuringCombat(),
			Effect:    effects.OldB823TargetGetsThenSacrificeThisWhenItLeaves("Heart Wolf", 2, "first strike"),
		}},
	})
}
