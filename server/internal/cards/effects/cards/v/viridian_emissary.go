package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viridian Emissary — Creature — Phyrexian Elf Scout {1}{G}, 2/1 (Commander
// 2015, S82 OLD #5546):
//
//	"When this creature dies, you may search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// A "you may search" (the prompt lets the controller decline the card and the
// shuffle) on a dies trigger, the land entering tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c93bce5a-48cc-4dc5-930b-1e084c2bae7d",
		Name:         "Viridian Emissary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Viridian Emissary — you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Optional:      true,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Viridian Emissary — search your library for a basic land card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
