package effects

// Wheel of Fortune — Sorcery {2}{R} (EDHREC rank 569):
//
//	"Each player discards their hand, then draws seven cards."
//
// The original wheel. Every discard happens before any draw, and
// each discard emits its own event, so discard payoffs queue while
// the spell is still resolving — Windfall's shape with a fixed seven
// instead of "the most anyone discarded". A player with an empty
// hand still draws seven.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8abd966-de7b-46a3-8ac7-8747ab35653a",
		Name:         "Wheel of Fortune",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The draws wait for a prompt a discard asks (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				for _, id := range tablePlayers(ctx) {
					if _, err := discardWholeHand(ctx.Game, id); err != nil {
						return err
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				for _, id := range tablePlayers(ctx) {
					if err := (DrawCards{Player: id, N: 7}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
