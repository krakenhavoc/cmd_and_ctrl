package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Proft, Consulting Detective — Legendary Creature — Human Detective
// {1}{U}, 2/2:
//
//	"Whenever you scry or surveil, you may pay {2}. If you do, put a
//	 +1/+1 counter on Proft and draw a card. (Draw after you scry or
//	 surveil.)"
//
// One ability watching EventScry and EventSurveil, each emitted once the
// scry or surveil is complete, so the card is drawn after it, as the
// reminder text says. The {2} is asked as the trigger resolves
// (MayPay). The counter goes on Proft only while he is still the same
// object on the battlefield; the draw happens regardless.
//
// No simplification.
func init() {
	const label = "Proft, Consulting Detective — you may pay {2} for a +1/+1 counter and a card"
	Register(Spec{
		OracleID:     "45473bdb-b96f-4213-a46d-b2dbdf6b5e56",
		Name:         "Proft, Consulting Detective",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventScry, game.EventSurveil}, ByYou, label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MayPay{
						Chooser:  ctx.Controller(),
						Cost:     "{2}",
						Question: "Proft, Consulting Detective — pay {2} to put a +1/+1 counter on Proft and draw a card?",
						OnPay: func(c *Context) error {
							if b09SourceStillOnBattlefield(c.Game, c.Item) {
								if err := (AddCounter{Target: c.Item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(c); err != nil {
									return err
								}
							}
							return DrawCards{Player: c.Controller(), N: 1}.Apply(c)
						},
					}.Apply(ctx)
				}),
		},
	})
}
