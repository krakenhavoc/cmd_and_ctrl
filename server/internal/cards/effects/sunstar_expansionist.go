package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "944a91ab-c217-4c1d-b920-2e5503c34546",
		Name:         "Sunstar Expansionist",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "an opponent controls more lands than you",
				Test: func(ctx *Context) bool { return b05OpponentControlsMoreLands(ctx.Game, ctx.Controller()) },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b05OpponentControlsMoreLands(g, source.Controller)
			}, "Sunstar Expansionist — create a Lander token",
				Do(CreateToken{Template: b16LanderToken(), N: 1}))),
			Landfall("Sunstar Expansionist — +1/+0 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(1, 0, "Sunstar Expansionist — +1/+0 until end of turn")),
		},
	})
}
