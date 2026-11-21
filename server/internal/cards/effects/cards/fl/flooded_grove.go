package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flooded Grove — Land (EDHREC rank 302):
//
//	"{T}: Add {C}.
//	 {G/U}, {T}: Add {G}{G}, {G}{U}, or {U}{U}."
//
// The Guildpact filter land, Simic half. See rugged_prairie.go for
// the shape's history and mechanics — this is the same body with
// Simic colours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc974eb4-72b9-4213-887b-8ee684b93420",
		Name:         "Flooded Grove",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{G/U}"},
				Produced: "{G|U}{G|U}",
				Label:    "{G/U}, {T}: Add {G}{G}, {G}{U}, or {U}{U}",
			},
		},
	})
}
