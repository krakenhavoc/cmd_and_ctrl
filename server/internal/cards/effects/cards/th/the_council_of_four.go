package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Council of Four — Legendary Creature — Human Noble {3}{W}{U}, 0/8
// (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"Whenever a player draws their second card during their turn, you draw a
//	 card.
//	 Whenever a player casts their second spell during their turn, you create a
//	 2/2 white Knight creature token."
//
// Bard the Bowman's second-card tally and Monk of the Open Hand's second-spell
// tally, read for ANY player (the event's actor) and only on that player's own
// turn: a draw or a cast in another player's turn is the first or second of
// nothing here, since the tally is the turn's. "You" is the Council's
// controller, whoever drew or cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3dfb006f-6620-4602-ac3c-d40a5a15a981",
		Name:         "The Council of Four",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventDrawCard && ev.Actor != uuid.Nil && g.Turn.ActiveSeat >= 0 &&
					g.Seats[g.Turn.ActiveSeat].ID == ev.Actor && g.TurnTallyFor(ev.Actor).CardsDrawn == 2
			}, "The Council of Four — you draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor != uuid.Nil && g.Turn.ActiveSeat >= 0 &&
					g.Seats[g.Turn.ActiveSeat].ID == ev.Actor && g.CastTallyFor(ev.Actor).Total == 2
			}, "The Council of Four — you create a 2/2 white Knight creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Knight"), N: 1})),
		},
	})
}
