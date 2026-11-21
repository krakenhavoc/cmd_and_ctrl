package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sungrass Prairie — Land:
//
//	"{1}, {T}: Add {G}{W}."
//
// The tap-and-a-tax dual, Selesnya half. See darkwater_catacombs.go for
// the shape's mechanics; this is Skycloud Expanse's body with green and
// white.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a28dff0-2bd6-4105-b73a-b6c4735833fd",
		Name:         "Sungrass Prairie",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{G}{W}",
				Label:    "{1}, {T}: Add {G}{W}",
			},
		},
	})
}
