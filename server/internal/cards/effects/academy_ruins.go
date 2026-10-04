package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a3da7d5b-2c2b-45fe-b9c5-413b8c8fc0a2",
		Name:         "Academy Ruins",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Put target artifact card from your graveyard on top of your library.",
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Targets: TargetCardInGraveyard("target artifact card from your graveyard", Artifact(), YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutChosenTargetOnTopOfLibrary(g, item)
			},
		}},
	})
}
