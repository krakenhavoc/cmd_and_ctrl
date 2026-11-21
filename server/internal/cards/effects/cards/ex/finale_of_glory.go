package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Finale of Glory — Sorcery {X}{W}{W} (EDHREC rank 2861):
//
//	"Create X 2/2 white Soldier creature tokens with vigilance. If X
//	 is 10 or more, also create X 4/4 white Angel creature tokens
//	 with flying and vigilance."
//
// White's mana sink. X rides the cast; the Soldiers are always made
// and the Angels join at ten. Both are their own templates with the
// printed colour and keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68273453-1059-4eab-8c27-5d96f998f3b1",
		Name:         "Finale of Glory",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			// The Angels wait for a prompt the Soldiers' creation asks (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("2/2 white Soldier with vigilance"), N: x}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				x := ctx.X()
				if x < 10 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("4/4 white Angel with flying and vigilance"), N: x}.Apply(ctx)
			}),
		),
	})
}
