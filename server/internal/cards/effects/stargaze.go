package effects

// Stargaze — Sorcery {X}{B}{B} (Secrets of Strixhaven Mystical Archive,
// #4834):
//
//	"Look at twice X cards from the top of your library. Put X cards from
//	 among them into your hand and the rest into your graveyard. You lose
//	 X life."
//
// Wail of the Forgotten's look-and-take with X on both counts: a private
// look at 2X cards, a mandatory pick of X of them (the prompt's floor and
// ceiling are both X), and the rest moved to the graveyard
// (takeRestGraveyard). The rulings: a library of fewer than 2X cards is
// looked at whole, and one of fewer than X cards goes into the hand whole
// (nothing is chosen for the graveyard); the life is lost either way.
//
// The life loss is its own step behind the pick (CR 608.2c), so it is not
// paid while the table is still choosing. At X = 0 the spell does nothing
// (XMatters).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db8b4869-f898-40ad-bbad-62390a02d1b1",
		Name:         "Stargaze",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				player := ctx.Controller()
				return TakeFromLibraryToHand{
					Player: player,
					Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 2*x),
					Max:    x,
					Label:  "Stargaze — put X cards into your hand and the rest into your graveyard",
					Key:    takeRestGraveyard,
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if x := ctx.X(); x > 0 {
					return GainLife{Amount: -x}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
