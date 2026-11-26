package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keeper of Progenitus — {3}{G} Creature — Elf Druid 1/3 (Shards of Alara,
// S82 OLD b513):
//
//	"Whenever a player taps a Mountain, Forest, or Plains for mana, that
//	 player adds one mana of any type that land produced."
//
// Mana Flare's triggered mana ability (CR 605.1b) narrowed to a land with one
// of three basic land types, read off the land as it was when it produced, so
// a land made a Forest counts and a Forest that lost the type does not. The three land words are read through Keeper's own text, so a
// text-changing effect reaches them (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a30241b4-5e92-44cd-b19b-7d6d59e95ada",
		Name:         "Keeper of Progenitus",
		Completeness: effects.CompletenessFull,
		ManaTriggers: []game.ManaTrigger{{
			Label: "Keeper of Progenitus — add one mana of any type that land produced",
			AppliesTo: func(prod game.ManaProduced, keeper *game.Card, _ *game.Game) bool {
				s := prod.Source
				return s.IsLand() && (s.HasSubtype(keeper.TextSubtype("Mountain")) ||
					s.HasSubtype(keeper.TextSubtype("Forest")) || s.HasSubtype(keeper.TextSubtype("Plains")))
			},
			Produced: effects.AddsOneManaOfAnyTypeProduced(),
		}},
	})
}
