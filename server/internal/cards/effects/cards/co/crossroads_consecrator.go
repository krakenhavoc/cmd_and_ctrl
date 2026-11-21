package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crossroads Consecrator — Creature — Human Cleric {G}, 1/2 (Eldritch
// Moon, #5541):
//
//	"{G}, {T}: Target attacking Human gets +1/+1 until end of turn."
//
// Infantry Veteran's pump at {G}, narrowed to an attacking Human.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4f41c7b-b4ae-46cd-9fdc-8b8e6c29a62d",
		Name:         "Crossroads Consecrator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Target attacking Human gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking Human", effects.AttackingCreature(), effects.HasSubtype("Human")),
			Effect:  effects.PumpAndGrantFirstTarget(1, 1, "Crossroads Consecrator — +1/+1 until end of turn"),
		}},
	})
}
