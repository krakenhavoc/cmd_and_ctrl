package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Priest of the Blessed Graf — Creature — Human Cleric {2}{W}, 1/2
// (Innistrad: Crimson Vow Commander):
//
//	"At the beginning of your end step, create X 1/1 white Spirit creature
//	 tokens with flying, where X is the number of opponents who control more
//	 lands than you."
//
// X is counted as the ability resolves, opponent by opponent (a player who
// has left the game is not an opponent); zero makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "726dbbce-cd5e-4605-aa1d-ae5ec80cc8cf",
		Name:         "Priest of the Blessed Graf",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Priest of the Blessed Graf — create X 1/1 white Spirits with flying, X is the number of opponents who control more lands than you",
				Do(CreateTokensForEach{
					Template: TokenCard("1/1 white Spirit with flying"),
					Name:     "opponents who control more lands than you",
					Count: func(ctx *Context) int {
						return vowB12OpponentsWithMoreLands(ctx.Game, ctx.Controller())
					},
				})),
		},
	})
}
