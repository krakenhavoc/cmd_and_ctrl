package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismatic Lens — Artifact {2}:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// The Signet shape one slot over: a free colourless tap and a paid
// fixing tap on the same permanent, `ManaAbilityCost.Mana` carrying
// the "{1}," component exactly as the Signet cycle's own does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3056f28-868b-401a-a528-7528d639bdeb",
		Name:         "Prismatic Lens",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
		},
	})
}
