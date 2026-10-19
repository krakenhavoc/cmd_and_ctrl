package effects

// Windfall — Sorcery for {2}{U}:
//
//	"Each player discards their hand, then draws cards equal to the
//	 greatest number of cards a player discarded this way."
//
// A symmetrical refill that isn't symmetrical at all in this deck:
// every card you pitch is a discard trigger, so Windfall is a
// Marauding Mako pump, a Magmakin volley and a fistful of Treasures
// before anyone draws anything.
//
// Order matters and the engine gets it right: all the discards
// happen first (so the draw count is fixed before any card is
// drawn), and each discard emits its own event, so the payoffs
// queue while the spell is still resolving.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08becc07-28bc-4a2f-a6b0-28a2998d2f50",
		Name:         "Windfall",
		Completeness: CompletenessFull,
		// The draws wait for a prompt a discard asks (a madness offer;
		// CR 608.2c), and read the count from the step record (#4163).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				most := 0
				for _, id := range tablePlayers(ctx) {
					n, err := discardWholeHand(ctx.Game, id)
					if err != nil {
						return err
					}
					most = max(most, n)
				}
				ctx.RecordStep(windfallRecord, nil, most)
				return nil
			}),
			Instruction(func(ctx *Context) error {
				most := ctx.StepRecord(windfallRecord).N()
				if most == 0 {
					return nil
				}
				for _, id := range tablePlayers(ctx) {
					if err := (DrawCards{Player: id, N: most}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}

// windfallRecord is the greatest number of cards a player discarded.
const windfallRecord = "windfall/most-discarded"
