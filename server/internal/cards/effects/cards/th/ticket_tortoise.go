package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ticket Tortoise — Artifact Creature — Turtle {2}, 3/1 (Aetherdrift,
// #4464):
//
//	"Defender
//	 When this creature enters, if an opponent controls more lands than
//	 you, you create a Treasure token."
//
// Dora Milaje Elite's intervening "if" (CR 603.4): the land count is read
// as the creature enters, so with no opponent ahead nothing triggers, and
// again as the trigger resolves (InterveningIf), so a land drop in
// response that evens the count makes it do nothing. Lands are counted
// through the layers and eliminated seats are skipped
// (b05OpponentControlsMoreLands). The Treasure is the predefined token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "952083c5-9ffa-4cd9-915f-b7de7f3e5326",
		Name:            "Ticket Tortoise",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "an opponent controls more lands than you",
				Test: func(ctx *effects.Context) bool {
					return effects.B05OpponentControlsMoreLands(ctx.Game, ctx.Controller())
				},
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && effects.B05OpponentControlsMoreLands(g, source.Controller)
			}, "Ticket Tortoise — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}))),
		},
	})
}
