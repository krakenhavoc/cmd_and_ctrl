package effects

// Wild Hypothesis — Sorcery {X}{G} (Secrets of Strixhaven, #4456):
//
//	"Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters
//	 on it.
//	 Surveil 2. (Look at the top two cards of your library, then put any number
//	 of them into your graveyard and the rest on top of your library in any
//	 order.)"
//
// The Fractal is Deekah's create-then-put shape (socB8FractalWithCountersSteps,
// written as a spell's ResolveSteps), with X as the count: the token is made
// first and the counters are then PUT on it, as printed. X is the number
// announced on cast (CR 601.2b, CR 107.3a). At X = 0 the Fractal is a 0/0 and
// dies to CR 704.5f, but the spell still surveils. The surveil is its own
// sentence and its own step (CR 608.2c): it is asked after the token and its
// counters exist, so the token is on the battlefield while the player decides.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e56507d-af77-47c8-a29a-410cf14cfe20",
		Name:         "Wild Hypothesis",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(append(
			socB8FractalWithCountersSteps("wild-hypothesis/fractal", func(ctx *Context) int { return ctx.X() }),
			Instruction(func(ctx *Context) error {
				return Surveil{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}),
		)...),
	})
}
