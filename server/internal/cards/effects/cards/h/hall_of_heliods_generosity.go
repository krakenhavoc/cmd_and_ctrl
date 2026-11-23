package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hall of Heliod's Generosity — Legendary Land:
//
//	"{T}: Add {C}.
//	 {1}{W}, {T}: Put target enchantment card from your graveyard on top
//	 of your library."
//
// Academy Ruins for enchantments and white. The target is an
// enchantment card in the controller's own graveyard, re-checked at
// resolution (CR 608.2b), so an enchantment exiled in response fizzles
// the ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fc070dc-f2f7-4648-8069-31d74790a39c",
		Name:         "Hall of Heliod's Generosity",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Put target enchantment card from your graveyard on top of your library.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target enchantment card from your graveyard", effects.Enchantment(), effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutChosenTargetOnTopOfLibrary(g, item)
			},
		}},
	})
}
