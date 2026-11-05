package effects

// Cryogenic Stasis — Instant {1}{U} (Star Trek Commander, #4974):
//
//	"Tap target creature and put a stun counter on it.
//	 Draw a card."
//
// Cryotheory Adept's tap-and-stun on the spell's target (the stun counter
// is CR 122.1d's untap replacement, which the engine applies), then the
// draw. A target gone by resolution counters the spell (CR 608.2b), so
// nothing is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "846ba66c-a4eb-4d00-a0f1-d9f9524b9695",
		Name:         "Cryogenic Stasis",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return tapTargetAndStunIt(ctx.Game, ctx.Item) }),
			DrawCards{N: 1},
		),
	})
}
