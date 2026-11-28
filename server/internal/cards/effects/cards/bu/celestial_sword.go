package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Celestial Sword — Artifact {6} (Ice Age, S82 OLD, #5636):
//
//	"{3}, {T}: Target creature you control gets +3/+3 until end of turn. Its
//	 controller sacrifices it at the beginning of the next end step."
//
// The sacrifice is a delayed trigger pinned to the creature as the object it
// is now (CR 603.7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f64c3d4-8ed6-43be-9583-33de4b7761fa",
		Name:         "Celestial Sword",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target creature you control gets +3/+3 until end of turn. Its controller sacrifices it at the beginning of the next end step.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  effects.OldB835TargetGetsThenSacrificeAtEnd(3, 3, "Celestial Sword — +3/+3"),
		}},
	})
}
