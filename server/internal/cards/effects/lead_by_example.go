package effects

// Lead by Example — Instant {1}{G} (Oath of the Gatewatch, #5545):
//
//	"Support 2. (Put a +1/+1 counter on each of up to two target
//	 creatures.)"
//
// Support from an instant may target any creatures (CR 701.41a):
// Shoulder to Shoulder's support, with no draw. With no target chosen
// the spell still resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a50418c-3cdc-4e3e-b02c-6da646093120",
		Name:         "Lead by Example",
		Completeness: CompletenessFull,
		Targets:      SupportTargets(2, false),
		OnResolve:    ResolveSteps(Instruction(func(ctx *Context) error { return SupportEffect(ctx.Game, ctx.Item) })),
	})
}
