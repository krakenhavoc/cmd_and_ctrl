package effects

// Path of the Pyromancer — Sorcery {4}{R} (March of the Machine
// Commander, #5648):
//
//	"Discard all the cards in your hand. Add {R} for each card discarded
//	 this way, then draw that many cards plus one.
//	 Will of the Planeswalkers — Starting with you, each player votes for
//	 planeswalk or chaos. If planeswalk gets more votes, planeswalk. If
//	 chaos gets more votes or the vote is tied, chaos ensues."
//
// The hand is discarded whole, as Tolarian Winds does it; the number
// discarded is filed on the resolution (a step record, so a restore taken
// while a discard asks something keeps it), and the mana and the draw
// read it. An empty hand discards nothing, adds nothing and draws one
// card. The mana is added while the spell resolves, so it is in the pool
// for the rest of the phase (CR 500.5). Then the vote
// (WillOfThePlaneswalkers); it happens at every table, and planeswalking
// and chaos do something only at a table playing Planechase.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b908fad9-25a6-441b-bba8-0d7fc12ee2dc",
		Name:         "Path of the Pyromancer",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				n, err := discardWholeHand(ctx.Game, ctx.Controller())
				ctx.RecordStep(pathOfThePyromancerRecord, nil, n)
				return err
			}),
			Instruction(func(ctx *Context) error {
				return AddMana{Amount: ManaN(ctx.StepRecord(pathOfThePyromancerRecord).N(), "R")}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return DrawCards{N: ctx.StepRecord(pathOfThePyromancerRecord).N() + 1}.Apply(ctx)
			}),
			WillOfThePlaneswalkers{Card: "Path of the Pyromancer"},
		),
	})
}

// pathOfThePyromancerRecord is the number of cards discarded.
const pathOfThePyromancerRecord = "path-of-the-pyromancer/discarded"
