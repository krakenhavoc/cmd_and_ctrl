package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faldorn, Dread Wolf Herald — Legendary Creature — Human Druid {1}{R}{G}, 3/3
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Whenever you cast a spell from exile or a land you control enters from
//	 exile, create a 2/2 green Wolf creature token.
//	 {1}, {T}, Discard a card: Exile the top card of your library. You may play
//	 it this turn."
//
// The trigger is Ghost-Spider's read of a cast from exile (EventCast stamps the
// origin in OldZone), widened from "plays a land" to any land that enters from
// exile (a zone move from exile to the battlefield, CR 400.7). The ability is
// Ghost-Spider's impulse draw (ExileTopWithPermission) behind a tap, a discard and
// {1}; the permission is "this turn" and covers a land play as well as a cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8e6294d-51cc-44ff-8bc6-b0a168e4f423",
		Name:         "Faldorn, Dread Wolf Herald",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventCast, game.EventZoneMove}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.GapB218CastFromExileOrLandEntersFromExile(ev, source, g)
			}, "Faldorn, Dread Wolf Herald — create a 2/2 green Wolf creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Discard a card: Exile the top card of your library. You may play it this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
