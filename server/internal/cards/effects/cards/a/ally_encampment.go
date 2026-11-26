package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ally Encampment — Land (Battle for Zendikar):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast an Ally
//	 spell.
//	 {1}, {T}, Sacrifice this land: Return target Ally you control to its
//	 owner's hand."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d293b69-12b7-4b50-a0a7-c4f493dee30b",
		Name:         "Ally Encampment",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Label:        "Add one mana of any color. Spend this mana only to cast an Ally spell",
				Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictSubtype("Ally")},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this land: Return target Ally you control to its owner's hand.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target Ally you control", effects.Subtype("Ally"), effects.YouControl()),
			Effect:  effects.BounceTheTargetToHand,
		}},
	})
}
