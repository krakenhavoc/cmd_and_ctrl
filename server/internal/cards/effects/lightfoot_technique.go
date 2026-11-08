package effects

// Lightfoot Technique — Instant {1}{W} (Tarkir: Dragonstorm, #4463):
//
//	"Put a +1/+1 counter on target creature. It gains flying and
//	 indestructible until end of turn."
//
// Fleeting Flight's first two sentences with indestructible beside the
// flying: the counter stays, and the keywords end in the cleanup step (CR
// 514.2). The keywords wait for a prompt the counter asks (CR 608.2c), and
// a target that has left by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "601b7481-19fc-4779-950e-883a9a22aec3",
		Name:         "Lightfoot Technique",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
			Instruction(func(ctx *Context) error {
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying", "indestructible"},
						Label: "Lightfoot Technique — flying and indestructible"}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
