package effects

// Reckless Spite — Instant {1}{B}{B}:
//
//	"Destroy two target nonblack creatures. You lose 5 life."
//
// Curtains' Call's destruction: one simultaneous event over the targets
// still legal as the spell resolves (CR 608.2b; a death trigger among them
// sees both). "Nonblack" is judged at announce and again at resolution. The
// life is lost whenever the spell resolves at all, so with one target gone
// it still costs 5; with both gone it fizzles and costs nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a684df3a-5441-4daa-86d1-c47a91b35e6a",
		Name:         "Reckless Spite",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("two target nonblack creatures", NonBlack()).WithCount(2, 2),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return destroyEachLegalTarget(ctx.Item, ctx) }),
			Instruction(func(ctx *Context) error {
				return GainLife{Player: ctx.Controller(), Amount: -5}.Apply(ctx)
			}),
		),
	})
}
