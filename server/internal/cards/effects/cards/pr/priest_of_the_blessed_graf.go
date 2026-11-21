package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "726dbbce-cd5e-4605-aa1d-ae5ec80cc8cf",
		Name:         "Priest of the Blessed Graf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Priest of the Blessed Graf — create X 1/1 white Spirits with flying, X is the number of opponents who control more lands than you",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("1/1 white Spirit with flying"),
					Name:     "opponents who control more lands than you",
					Count: func(ctx *effects.Context) int {
						return effects.VowB12OpponentsWithMoreLands(ctx.Game, ctx.Controller())
					},
				})),
		},
	})
}
