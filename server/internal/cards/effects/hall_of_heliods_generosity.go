package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2fc070dc-f2f7-4648-8069-31d74790a39c",
		Name:         "Hall of Heliod's Generosity",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}: Put target enchantment card from your graveyard on top of your library.",
			Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
			Targets: TargetCardInGraveyard("target enchantment card from your graveyard", Enchantment(), YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutChosenTargetOnTopOfLibrary(g, item)
			},
		}},
	})
}
