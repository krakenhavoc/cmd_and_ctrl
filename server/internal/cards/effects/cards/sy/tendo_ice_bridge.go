package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tendo Ice Bridge — Land (Betrayers of Kamigawa, #5602):
//
//	"This land enters with a charge counter on it.
//	 {T}: Add {C}.
//	 {T}, Remove a charge counter from this land: Add one mana of any
//	 color."
//
// A vivid land with one counter that enters untapped: the counter is a
// self-replacement on the land's own entry (b10EntersWithCounters), and
// the second mana ability takes the vivid lands' RemoveCounters cost, so
// with no counter left it is not offered and the land taps for {C} only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e006a4b-8dde-4416-8cb4-8401562d0fd5",
		Name:         "Tendo Ice Bridge",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterCharge, 1, "Tendo Ice Bridge: enters with a charge counter"),
		},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: effects.ManaAbilityCost{
					Tap:            true,
					RemoveCounters: effects.RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters,
				},
				Produced: "{W|U|B|R|G}",
				Label:    "Remove a charge counter: Add one mana of any color",
			},
		},
	})
}
