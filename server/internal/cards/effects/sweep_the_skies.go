package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sweep the Skies — Sorcery {X}{U}{U} (Modern Horizons 2, #5501):
//
//	"Converge — Create a 1/1 colorless Thopter artifact creature token
//	 with flying for each color of mana spent to cast this spell."
//
// Radiant Flames' converge count (ctx.ColorsSpentCount, the distinct
// colours of the payment), so {U}{U} with X=0 still makes one Thopter; X
// only lets the caster spend more colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "7596fe49-c9cf-4415-ac1b-156cbbf9604e",
		Name:                "Sweep the Skies",
		Completeness:        CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Purpose:             game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := ctx.ColorsSpentCount()
			if n <= 0 {
				return nil
			}
			return CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: n}.Apply(ctx)
		},
	})
}
