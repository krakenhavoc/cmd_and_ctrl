package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Late to Dinner — Sorcery {3}{W}:
//
//	"Return target creature card from your graveyard to the
//	 battlefield. Create a Food token."
//
// Zombify in white with a Food stapled on. The Food is not
// conditional on the reanimation, so the token creation sits outside
// the ok check rather than inside it.
//
// The one case where you get neither is the spell not resolving at
// all: this has a single target, so a target that became illegal
// between announce and resolution takes the whole spell with it (CR
// 608.2b) and OnResolve never runs. That is the printed card, not a
// simplification — Late to Dinner is uncastable with an empty
// graveyard for the same reason.
func init() {
	Register(Spec{
		OracleID: "31bf199b-dfb1-428e-96a5-eb25104e2b43",
		Name:     "Late to Dinner",
		Targets:  targetCreatureInYourGraveyard(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, _ = reanimateSingleTarget(ctx, ctx.Controller())
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   FoodToken(),
				N:          1,
			}.Apply(ctx)
		},
	})
}
