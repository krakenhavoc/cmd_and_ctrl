package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimmering Grotto — Land (Lorwyn, #5593):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// Conduit Pylons' two mana abilities without the surveil: a colourless tap,
// and a filter that turns one mana into any colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bae49475-fe01-400b-8959-f0dde959577c",
		Name:         "Shimmering Grotto",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.PainlessColorless(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
	})
}
