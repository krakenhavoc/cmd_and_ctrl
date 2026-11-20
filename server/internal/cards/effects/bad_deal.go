package effects

// Bad Deal — Sorcery {4}{B}{B} (Core Set 2021, #5511):
//
//	"You draw two cards and each opponent discards two cards. Each player
//	 loses 2 life."
//
// The draw, then Bite of the Black Rose's each-opponent discard, then
// Crushing Disappointment's table-wide life loss (eachPlayerLosesTwoLife).
// The life waits for the discards the opponents are asked about (CR
// 608.2c, ResolveSteps).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47258c3e-b4a7-466d-9464-3924f75d9293",
		Name:         "Bad Deal",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if err := (DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
					return err
				}
				eachOpponentDiscards(ctx, 2)
				return nil
			}),
			Instruction(eachPlayerLosesTwoLife),
		),
	})
}
