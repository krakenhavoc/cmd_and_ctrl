package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "68273453-1059-4eab-8c27-5d96f998f3b1",
		Name:         "Finale of Glory",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			if err := (CreateToken{Controller: item.Controller, Template: TokenCard("2/2 white Soldier with vigilance"), N: x}).Apply(ctx); err != nil {
				return err
			}
			if x < 10 {
				return nil
			}
			return CreateToken{Controller: item.Controller, Template: TokenCard("4/4 white Angel with flying and vigilance"), N: x}.Apply(ctx)
		},
	})
}
