package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Form a Posse — Sorcery {X}{R}{W} (Outlaws of Thunder Junction, #4468):
//
//	"Create X 1/1 red Mercenary creature tokens with '{T}: Target creature
//	 you control gets +1/+0 until end of turn. Activate only as a
//	 sorcery.'"
//
// The set's shared Mercenary token (MercenaryToken), X of them at once.
// X is the value announced as the spell was cast (CR 107.3); with X = 0
// nothing is created.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8923462f-0540-41e8-931e-97b5f1dadd51",
		Name:         "Form a Posse",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.X() <= 0 {
				return nil
			}
			return CreateToken{Template: MercenaryToken(), N: ctx.X()}.Apply(ctx)
		},
	})
}
