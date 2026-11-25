package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saruli Caretaker — Creature — Dryad {G}, 0/3 (Ravnica Allegiance):
//
//	"Defender
//	 {T}, Tap an untapped creature you control: Add one mana of any color."
//
// Loam Dryad's mana ability: tapping the Caretaker and another untapped
// creature are both costs, so it is a mana ability that cannot be activated
// with no other creature to tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e2736e6-b812-43a4-bd79-6f6095344fe1",
		Name:            "Saruli Caretaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, TapOthers: effects.TapAnotherUntapped("an untapped creature you control", effects.Creature()).TapOthers},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
