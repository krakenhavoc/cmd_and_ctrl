package effects

// Decode Transmissions — Sorcery {2}{B} (Edge of Eternities):
//
//	"You draw two cards and lose 2 life.
//	 Void — If a nonland permanent left the battlefield this turn or a
//	 spell was warped this turn, instead you draw two cards and each
//	 opponent loses 2 life."
//
// "Instead" swaps the second half only: you draw two cards either way, and
// void (void.go), read as the spell resolves, decides whether the loss is
// yours or every opponent's. Two steps, so the loss waits for a prompt the
// draw asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42510758-6167-420b-bb70-65f914b7a259",
		Name:         "Decode Transmissions",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			DrawCards{N: 2},
			Instruction(func(ctx *Context) error {
				if Void(ctx.Game) {
					return EachOpponentLosesLife{N: 2}.Apply(ctx)
				}
				return youLoseLife(2)(ctx)
			}),
		),
	})
}
