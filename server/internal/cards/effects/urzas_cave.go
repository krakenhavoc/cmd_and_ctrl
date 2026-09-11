package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza's Cave — Land — Urza's Cave (EDHREC rank 451):
//
//	"{T}: Add {C}.
//	 {3}, {T}, Sacrifice this land: Search your library for a land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// A colourless land that becomes ANY land — the Cabal Coffers deck's
// third tutor for it, the Field of the Dead deck's landfall trigger.
// Myriad Landscape's cost shape with a wider predicate: any land
// card, not only basics, which is why it costs one more.
//
// The fetched land arrives tapped by the printed text (TappedOnEntry)
// and, since #263, also runs its own entry replacements.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4474ecee-0ec3-409b-90df-738d9313fe3c",
		Name:         "Urza's Cave",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}, Sacrifice this land: Search your library for a land card, put it onto the battlefield tapped, then shuffle.",
			Cost:  Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     b03IsLandCard,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Urza's Cave — a land card, onto the battlefield tapped",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
