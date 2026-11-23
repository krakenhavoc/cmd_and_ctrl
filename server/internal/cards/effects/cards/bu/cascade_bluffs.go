package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cascade Bluffs — Land (EDHREC rank 294):
//
//	"{T}: Add {C}.
//	 {U/R}, {T}: Add {U}{U}, {U}{R}, or {R}{R}."
//
// The Guildpact filter land, Izzet half. See rugged_prairie.go for
// the shape's history and mechanics — this is the same body with
// Izzet colours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1603384-4361-49c9-98aa-7785fc3504c4",
		Name:         "Cascade Bluffs",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{U/R}"},
				Produced: "{U|R}{U|R}",
				Label:    "{U/R}, {T}: Add {U}{U}, {U}{R}, or {R}{R}",
			},
		},
	})
}
