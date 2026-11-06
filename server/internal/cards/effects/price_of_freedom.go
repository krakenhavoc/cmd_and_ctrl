package effects

// Price of Freedom — Sorcery — Lesson {1}{R}:
//
//	"Destroy target artifact or land an opponent controls. Its
//	 controller may search their library for a basic land card, put it
//	 onto the battlefield tapped, then shuffle.
//	 Draw a card."
//
// Avengers Disassembled's land bullet over an artifact or land, then a
// draw. The search is optional, and a player who declines does not
// shuffle (the ruling); SearchLibrary's Optional does both. The draw
// waits for the search's answer (CR 608.2c), so it is a second step.
func init() {
	Register(Spec{
		OracleID:     "5e6faa98-cf55-430d-87ce-8085f68aa20a",
		Name:         "Price of Freedom",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or land an opponent controls", Or(Artifact(), Land()), OpponentControls()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return destroyThenItsControllerMaySearchForATappedBasic(ctx, id,
					"Price of Freedom — you may search for a basic land")
			}),
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
