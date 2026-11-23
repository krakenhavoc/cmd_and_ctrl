package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunstar Expansionist — Creature — Human Knight {1}{W}, 2/3 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, if an opponent controls more lands than
//	 you, create a Lander token.
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +1/+0 until end of turn."
//
// The enters trigger is Dora Milaje Elite's intervening "if" (CR 603.4):
// checked as the creature enters and again as the trigger resolves, so a
// land played in response that evens the count makes no Lander. The
// landfall pump is Remnant Elemental's, +1/+0 until the cleanup step
// (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "944a91ab-c217-4c1d-b920-2e5503c34546",
		Name:         "Sunstar Expansionist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "an opponent controls more lands than you",
				Test: func(ctx *effects.Context) bool {
					return effects.B05OpponentControlsMoreLands(ctx.Game, ctx.Controller())
				},
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && effects.B05OpponentControlsMoreLands(g, source.Controller)
			}, "Sunstar Expansionist — create a Lander token",
				effects.Do(effects.CreateToken{Template: effects.B16LanderToken(), N: 1}))),
			effects.Landfall("Sunstar Expansionist — +1/+0 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(1, 0, "Sunstar Expansionist — +1/+0 until end of turn")),
		},
	})
}
