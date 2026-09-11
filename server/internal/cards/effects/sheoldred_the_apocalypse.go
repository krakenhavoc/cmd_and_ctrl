package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sheoldred, the Apocalypse — Legendary Creature — Phyrexian Praetor
// {2}{B}{B}, 4/5 (EDHREC rank 458):
//
//	"Deathtouch
//	 Whenever you draw a card, you gain 2 life.
//	 Whenever an opponent draws a card, they lose 2 life."
//
// The most-played four-drop of its era: every draw step at the table
// is a two-point swing. Two printed abilities, two triggers, both on
// EventDrawCard (which fires once per card, so a Divination is two
// triggers): one gated to the controller's own draws, one to
// everyone else's. The opponent's ID is captured in Build by value
// — ev.Actor, the Edric pattern — so the Effect drains the player
// who drew rather than whoever happens to be targeted. Life loss,
// not damage, so nothing prevents it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "34f34409-326d-4994-a0ea-1a69aa278f03",
		Name:            "Sheoldred, the Apocalypse",
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Sheoldred, the Apocalypse — you gain 2 life",
						func(g *game.Game, item *game.StackItem) error {
							return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor != uuid.Nil && ev.Actor != source.Controller
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					drawer := ev.Actor
					return game.NewTriggeredItem(source, "Sheoldred, the Apocalypse — an opponent loses 2 life",
						func(g *game.Game, item *game.StackItem) error {
							if g.PlayerByIDForEffect(drawer) == nil {
								return nil
							}
							return g.ChangePlayerLifeForEffect(item.SourceCardID, drawer, -2)
						})
				},
			},
		},
	})
}
