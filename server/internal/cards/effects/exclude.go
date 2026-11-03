package effects

// Exclude — Instant {2}{U}:
//
//	"Counter target creature spell.
//	 Draw a card."
//
// Essence Scatter's target with a cantrip. The picker offers only
// creature spells and the engine refuses a noncreature pick (CR 601.2c).
// The draw follows the counter as a second step, so it waits for any
// prompt the counter asks (a countered commander's zone choice). If the
// target spell has left the stack by resolution, Exclude does nothing at
// all, draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8c9f91a-b1e7-451d-b6db-ae865e2b853c",
		Name:         "Exclude",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if len(ctx.Item.Targets) == 0 {
					return nil
				}
				return CounterTarget{StackID: ctx.Item.Targets[0].ID}.Apply(ctx)
			}),
			DrawCards{N: 1},
		),
	})
}
