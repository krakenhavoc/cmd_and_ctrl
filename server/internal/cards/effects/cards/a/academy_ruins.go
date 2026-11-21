package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Academy Ruins — Legendary Land:
//
//	"{T}: Add {C}.
//	 {1}{U}, {T}: Put target artifact card from your graveyard on top of
//	 your library."
//
// Mistveil Plains's shape with the library end and the card filter
// swapped: the put goes through the shared exit primitive, so a
// commander card in the graveyard gets the CR 903.9 offer on the way to
// the library, and a target that leaves the graveyard in response makes
// the ability fizzle (CR 608.2b). The Legendary supertype is read off
// the printed TypeLine.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3da7d5b-2c2b-45fe-b9c5-413b8c8fc0a2",
		Name:         "Academy Ruins",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Put target artifact card from your graveyard on top of your library.",
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutChosenTargetOnTopOfLibrary(g, item)
			},
		}},
	})
}
