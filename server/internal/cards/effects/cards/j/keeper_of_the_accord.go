package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Keeper of the Accord — Creature — Human Soldier {3}{W}, 3/4 (EDHREC
// rank 1261):
//
//	"At the beginning of each opponent's end step, if that player
//	 controls more creatures than you, create a 1/1 white Soldier
//	 creature token.
//	 At the beginning of each opponent's end step, if that player
//	 controls more lands than you, you may search your library for a
//	 basic Plains card, put it onto the battlefield tapped, then
//	 shuffle."
//
// White's catch-up engine: two separate end-step triggers, each with
// an intervening-if against the player whose end step it is. Both
// watch EventBeginEndStep for an actor other than the controller —
// every other seat is an opponent — and re-check their condition at
// resolution as well as at trigger time (CR 603.4), reading the
// player whose end step it was off the item's carried trigger context
// (item.Trigger.Event.Actor, #1223). The land search is "you may", so
// the S22 chooser can decline it and the shuffle with it; the Soldier
// is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49b9a3ac-d245-431a-bf40-64d1b8dc2b91",
		Name:         "Keeper of the Accord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
						effects.B11OpponentControlsMoreThanYou(g, source.Controller, ev.Actor, game.Card.IsCreature)
				},
				Key: "Keeper of the Accord — create a Soldier",
				Effect: effects.Do(effects.If{
					Cond: effects.CondBoard{Name: "that player controls more creatures than you", Test: func(ctx *effects.Context) bool {
						item := ctx.Item
						return item.Trigger != nil &&
							effects.B11OpponentControlsMoreThanYou(ctx.Game, item.Controller, item.Trigger.Event.Actor, game.Card.IsCreature)
					}},
					Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}},
				}),
			},
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
						effects.B11OpponentControlsMoreThanYou(g, source.Controller, ev.Actor, game.Card.IsLand)
				},
				Key: "Keeper of the Accord — you may search for a basic Plains",
				Effect: func(g *game.Game, item *game.StackItem) error {
					opp := item.Trigger.Event.Actor
					if !effects.B11OpponentControlsMoreThanYou(g, item.Controller, opp, game.Card.IsLand) {
						return nil
					}
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return effects.IsBasicLand(c) && c.HasSubtypeWord("Plains")
						},
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						TappedOnEntry: true,
						Optional:      true,
						Reason:        "Keeper of the Accord — a basic Plains card, onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
