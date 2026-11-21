package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Circuitous Route — Sorcery {3}{G} (EDHREC rank 1767):
//
//	"Search your library for up to two basic land cards and/or Gate
//	 cards, put them onto the battlefield tapped, then shuffle."
//
// Explosive Vegetation for the Gates deck. Blighted Woodland's search
// with a wider predicate: a basic land, or any land with the Gate
// subtype (a Guildgate, Maze's End's friends). The S22 chooser picks
// up to two; "up to" is the chooser declining a slot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30afacd9-4680-4aac-8c22-584f9418822d",
		Name:         "Circuitous Route",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player: item.Controller,
				Predicate: func(c game.Card) bool {
					return effects.IsBasicLand(c) || (c.IsLand() && c.HasSubtype("Gate"))
				},
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Circuitous Route — up to two basic lands and/or Gates, onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
