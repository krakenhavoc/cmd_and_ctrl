package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archaeological Dig — Land (Invasion, S82 older sets batch 663, #5616):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Add one mana of any color."
//
// Crystal Vein's pair of mana abilities, the cash-in making one mana of a color
// the controller picks (the any-color pipe, which lists their commander
// identity first). Both resolve without the stack (CR 605.3b); the land is
// sacrificed as part of the second one's cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf438848-da86-4db6-b3b8-4dd8570be3b8",
		Name:         "Archaeological Dig",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}, Sacrifice this land: Add one mana of any color",
			},
		},
	})
}
