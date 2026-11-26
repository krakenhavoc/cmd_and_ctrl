package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quicksand — Land (Conspiracy):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Target attacking creature without flying gets
//	 -1/-2 until end of turn."
//
// "Without flying" is part of the target's legality, read through the layered
// view, so a creature that has flying cannot be chosen.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef2bb4fa-f292-4d19-aaa4-cfbe445caf45",
		Name:         "Quicksand",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Target attacking creature without flying gets -1/-2 until end of turn.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target attacking creature without flying", effects.And(effects.IsAttackingCreature(), effects.WithoutKeyword("flying"))),
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -2)),
			Effect:  effects.PumpAndGrantFirstTarget(-1, -2, "Quicksand — -1/-2 until end of turn"),
		}},
	})
}
