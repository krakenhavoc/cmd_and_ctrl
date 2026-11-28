package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Centaur Rootcaster — Creature — Centaur Druid {3}{G}, 2/2 (Judgment):
//
//	"Whenever this creature deals combat damage to a player, you may search
//	 your library for a basic land card, put that card onto the battlefield
//	 tapped, then shuffle."
//
// Mwonvuli Acid-Moss's tapped search for a land card, narrowed to a basic land
// and made optional ("you may"): the player may find nothing, and the library
// is shuffled either way.
//
// No simplification.
func init() {
	const name = "Centaur Rootcaster"
	effects.Register(effects.Spec{
		OracleID:     "0aaefa00-86f4-4e12-87ae-3a362d862fa4",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(name+" — you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Optional:      true,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        name + " — a basic land card, onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
