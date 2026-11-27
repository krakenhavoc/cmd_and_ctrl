package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vernal Bloom — Enchantment {3}{G} (Eighth Edition, #5607):
//
//	"Whenever a Forest is tapped for mana, its controller adds an
//	 additional {G}."
//
// A triggered mana ability (CR 605.1b): it adds the mana as the Forest is
// tapped, without the stack, to the player who tapped it, whoever that is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "238b270c-152a-48e5-a955-65f5cb80c174",
		Name:         "Vernal Bloom",
		Completeness: effects.CompletenessFull,
		ManaTriggers: []game.ManaTrigger{{
			Label: "Vernal Bloom — its controller adds an additional {G}",
			AppliesTo: func(prod game.ManaProduced, source *game.Card, _ *game.Game) bool {
				return prod.Source.IsLand() && prod.Source.HasSubtype(source.TextSubtype("Forest"))
			},
			Produced: effects.AddsFixedMana("{G}"),
		}},
	})
}
