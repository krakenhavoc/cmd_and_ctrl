package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4474ecee-0ec3-409b-90df-738d9313fe3c",
		Name:         "Urza's Cave",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this land: Search your library for a land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.B03IsLandCard,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Urza's Cave — a land card, onto the battlefield tapped",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
