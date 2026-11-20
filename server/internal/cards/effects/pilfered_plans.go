package effects

// Pilfered Plans — Sorcery {1}{U}{B}:
//
//	"Target player mills two cards. Draw two cards."
//
// The shared TargetPlayerMills body for two cards, then the draw. The draw
// waits for any prompt the mill asks (CR 608.2c, #4155).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b734cfd9-afef-41f1-8377-68c885fb3a69",
		Name:         "Pilfered Plans",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return TargetPlayerMills(2)(ctx.Game, ctx.Item) }),
			DrawCards{N: 2},
		),
	})
}
