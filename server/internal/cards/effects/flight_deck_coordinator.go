package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flight-Deck Coordinator — Creature — Human Soldier {2}{W}, 3/3 (Edge of
// Eternities, #4461):
//
//	"At the beginning of your end step, if you control two or more tapped
//	 creatures, you gain 2 life."
//
// Vaultguard Trooper's end-step trigger with a life gain in place of the
// discard. The "if" is an intervening if (CR 603.4): checked as the end
// step begins and again as the ability resolves, so untapping a creature
// in response leaves it doing nothing. The Coordinator itself counts
// toward the two tapped creatures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57927909-11a9-44ef-989d-1dcca1f6e23f",
		Name:         "Flight-Deck Coordinator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control two or more tapped creatures",
				Test: func(ctx *Context) bool { return b24TappedCreaturesControlled(ctx.Game, ctx.Controller()) >= 2 },
			}, On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24TappedCreaturesControlled(g, source.Controller) >= 2
			}), "Flight-Deck Coordinator — you gain 2 life", Do(GainLife{Amount: 2}))),
		},
	})
}
