package effects

// Fractal Anomaly — Instant {U} (Secrets of Strixhaven, #4456):
//
//	"Create a 0/0 green and blue Fractal creature token and put X +1/+1
//	 counters on it, where X is the number of cards you've drawn this turn."
//
// The Fractal is Deekah's create-then-put shape (socB8FractalWithCountersSteps,
// written as a spell's ResolveSteps). X is read as the spell resolves (CR
// 608.2c) off your per-turn tally of cards drawn, which counts every card you
// drew this turn however you drew it, including before the spell was cast.
// With none drawn the Fractal is a 0/0 and dies to CR 704.5f.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "644379e7-782e-48e1-9b47-c4dc62e62d85",
		Name:         "Fractal Anomaly",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(socB8FractalWithCountersSteps("fractal-anomaly/fractal", func(ctx *Context) int {
			return ctx.Game.TurnTallyFor(ctx.Controller()).CardsDrawn
		})...),
	})
}
