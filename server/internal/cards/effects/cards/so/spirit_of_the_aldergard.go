package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirit of the Aldergard — Snow Creature — Bear Spirit {3}{G}, 0/4 (Kaldheim):
//
//	"When this creature enters, search your library for a snow land card,
//	 reveal it, put it into your hand, then shuffle.
//	 This creature gets +1/+0 for each other snow permanent you control."
//
// The pump is a layer 7c per-count static read from the board on every layer
// pass, so it follows a snow permanent entering, leaving or changing hands;
// the Spirit does not count itself ("other").
//
// No simplification.
func init() {
	pump := effects.PumpSelfCreatureOrAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
		return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool {
			return c.InstanceID != source.InstanceID && c.HasSupertype("snow")
		})
	})
	pump.Label = "This creature gets +1/+0 for each other snow permanent you control."
	effects.Register(effects.Spec{
		OracleID:     "ff3bc67b-f435-4bb7-9c21-8d24d8808c1a",
		Name:         "Spirit of the Aldergard",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Static:       []game.StaticAbility{pump},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spirit of the Aldergard — search your library for a snow land card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsLand() && c.HasSupertype("snow") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Spirit of the Aldergard — a snow land card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
