package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal Quarry — Land:
//
//	"{T}: Add {C}.
//	 {5}, {T}: Add {W}{U}{B}{R}{G}."
//
// Two mana abilities, the second a fixed five-colour string behind a
// {5} cost (Prismatic Lens's costed ability, with no choice to make).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca68648f-fe3a-4770-9842-a3dc2310f099",
		Name:         "Crystal Quarry",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{5}"},
				Produced: "{W}{U}{B}{R}{G}",
				Label:    "{5}, {T}: Add {W}{U}{B}{R}{G}",
			},
		},
	})
}
