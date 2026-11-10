package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Form a Posse — Sorcery {X}{R}{W} (Outlaws of Thunder Junction, #4468):
//
//	"Create X 1/1 red Mercenary creature tokens with '{T}: Target creature
//	 you control gets +1/+0 until end of turn. Activate only as a sorcery.'"
//
// X Mercenaries from the set's shared template (MercenaryToken). X is the
// announced value; X = 0 makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8923462f-0540-41e8-931e-97b5f1dadd51",
		Name:         "Form a Posse",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: MercenaryToken(), N: ctx.X()}.Apply(ctx)
		},
	})
}
