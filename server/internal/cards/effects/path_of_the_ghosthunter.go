package effects

// Path of the Ghosthunter — Sorcery {X}{1}{W} (March of the Machine
// Commander, #5648):
//
//	"Create X 1/1 white Spirit creature tokens with flying.
//	 Will of the Planeswalkers — Starting with you, each player votes for
//	 planeswalk or chaos. If planeswalk gets more votes, planeswalk. If
//	 chaos gets more votes or the vote is tied, chaos ensues."
//
// X rides the cast. With X 0 no token is made and the vote still happens
// (WillOfThePlaneswalkers), after anything the tokens' entry asks (CR
// 608.2c). The vote happens at every table; planeswalking and chaos do
// something only at a table playing Planechase.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b79c3ae8-ac25-4a16-a981-a41721040fad",
		Name:         "Path of the Ghosthunter",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				return CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: x}.Apply(ctx)
			}),
			WillOfThePlaneswalkers{Card: "Path of the Ghosthunter"},
		),
	})
}
