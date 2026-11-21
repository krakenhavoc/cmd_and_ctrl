package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eiganjo Uprising — Sorcery {X}{R}{W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Create X 2/2 white Samurai creature tokens with vigilance. They gain
//	 menace and haste until end of turn.
//	 Each opponent creates X minus one 2/2 white Samurai creature tokens with
//	 vigilance."
//
// Only the caster's tokens gain menace and haste (Mardu Monument's keyed tail
// over the tokens the creation made, so a paused creation is waited for). Each
// opponent then makes X minus one, none at X of 0 or 1 (CR 107.1b: a negative
// number is treated as zero). Opponents' tokens are theirs, with none of the
// keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe2fe9f2-908b-49a9-9648-b01e0417a364",
		Name:         "Eiganjo Uprising",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x > 0 {
				tail := ctx.ItemTail(effects.ThenCreatedGainKeywordUntilEOT)
				tail.Word = "menace"
				tail.Words = []string{"haste"}
				if err := ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
					Controller: ctx.Controller(),
					Source:     item.SourceCardID,
					Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.NeoB08SamuraiToken()), Count: x}},
				}, tail); err != nil {
					return err
				}
			}
			if x < 2 {
				return nil
			}
			for _, opp := range ctx.Opponents() {
				if err := (effects.CreateToken{Controller: opp, Template: effects.NeoB08SamuraiToken(), N: x - 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
