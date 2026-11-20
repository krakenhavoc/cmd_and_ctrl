package effects

// Wicked Pact — Sorcery {1}{B}{B} (Portal, #5630):
//
//	"Destroy two target nonblack creatures. You lose 5 life."
//
// Reckless Spite's text, word for word. The life is lost even if one
// target has become illegal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39e21a5a-b278-478a-854c-17695c0f6246",
		Name:         "Wicked Pact",
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
