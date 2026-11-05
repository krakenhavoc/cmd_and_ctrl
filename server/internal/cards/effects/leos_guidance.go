package effects

// Leo's Guidance — Instant {2}{W} (Teenage Mutant Ninja Turtles
// Commander, #5011):
//
//	"Put a +1/+1 counter on each of up to three target creatures. Untap
//	 them."
//
// Up to three creature targets, zero a legal choice. Each target still
// legal as the spell resolves (CR 608.2b) gets a counter, and then those
// same targets are untapped: the untap waits for any prompt the counters
// ask (a CR 616 order).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f386bca6-a08e-462b-b1d9-c3b03929c68d",
		Name:         "Leo's Guidance",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return putPlusOneCounterOnEachLegalTarget(ctx.Game, ctx.Item) }),
			Instruction(func(ctx *Context) error { return UntapEachLegalTarget(ctx.Game, ctx.Item) }),
		),
	})
}
