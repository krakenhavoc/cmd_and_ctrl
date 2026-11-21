package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// School of the Unseen — Land (Alliances, #5634):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Add one mana of any color."
//
// Two mana abilities (CR 605.1a). The second pays {2} from the pool
// before the land taps, and the colour is one pick among the five
// (CR 106.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5028dfe8-c505-4643-b493-760b1f19d47f",
		Name:         "School of the Unseen",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.ColorlessTapRow(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{2}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{2}, {T}: Add one mana of any color.",
			},
		},
	})
}
