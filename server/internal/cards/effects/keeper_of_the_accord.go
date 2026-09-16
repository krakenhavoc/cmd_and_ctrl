package effects

import (
	"github.com/google/uuid"

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
// resolution as well as at trigger time (CR 603.4). The land search
// is "you may", so the S22 chooser can decline it and the shuffle
// with it; the Soldier is not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49b9a3ac-d245-431a-bf40-64d1b8dc2b91",
		Name:         "Keeper of the Accord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
						b11OpponentControlsMoreThanYou(g, source.Controller, ev.Actor, game.Card.IsCreature)
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					opp := ev.Actor
					return game.NewTriggeredItem(source, "Keeper of the Accord — create a Soldier",
						func(g *game.Game, item *game.StackItem) error {
							if !b11OpponentControlsMoreThanYou(g, item.Controller, opp, game.Card.IsCreature) {
								return nil
							}
							return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Soldier"), N: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
						b11OpponentControlsMoreThanYou(g, source.Controller, ev.Actor, game.Card.IsLand)
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					opp := ev.Actor
					return game.NewTriggeredItem(source, "Keeper of the Accord — you may search for a basic Plains",
						func(g *game.Game, item *game.StackItem) error {
							if !b11OpponentControlsMoreThanYou(g, item.Controller, opp, game.Card.IsLand) {
								return nil
							}
							return SearchLibrary{
								Player: item.Controller,
								Predicate: func(c game.Card) bool {
									return IsBasicLand(c) && c.HasSubtype("Plains")
								},
								Dest:          game.ZoneBattlefield,
								Limit:         1,
								Reveal:        true,
								Shuffle:       true,
								TappedOnEntry: true,
								Optional:      true,
								Reason:        "Keeper of the Accord — a basic Plains card, onto the battlefield tapped",
							}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
