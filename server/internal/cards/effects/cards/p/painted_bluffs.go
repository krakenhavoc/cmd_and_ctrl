package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Painted Bluffs — Land — Desert:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// A mana ability with a mana cost, the way the Signets pay {1}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b66deeb5-7371-4f06-b10e-d65165bc07b2",
		Name:         "Painted Bluffs",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.PainlessColorless(),
			{
				Cost:     effects.ManaAbilityCost{Mana: "{1}", Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
	})
}
