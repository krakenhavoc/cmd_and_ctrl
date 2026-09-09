package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
func init() {
	Register(Spec{
		OracleID: "08becc07-28bc-4a2f-a6b0-28a2998d2f50",
		Name:     "Windfall",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			players := tablePlayers(ctx)
			most := 0
			for _, id := range players {
				n, err := discardWholeHand(ctx.Game, id)
				if err != nil {
					return err
				}
				if n > most {
					most = n
				}
			}
			if most == 0 {
				return nil
			}
			for _, id := range players {
				if err := (DrawCards{Player: id, N: most}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
