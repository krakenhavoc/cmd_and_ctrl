package effects

// Unexplained Disappearance — Instant {1}{U} (Guilds of Ravnica, #5525):
//
//	"Return target creature to its owner's hand.
//	 Surveil 1."
//
// Two steps: the shared bounceTheTarget, then surveil 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "18d16129-da25-4120-a343-64ce962abba2",
		Name:         "Unexplained Disappearance",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return bounceTheTarget(nil, ctx) }),
			Surveil{N: 1},
		),
	})
}
