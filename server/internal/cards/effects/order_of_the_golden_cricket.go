package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Order of the Golden Cricket — Creature — Kithkin Knight {1}{W}, 2/2
// (Morningtide, #5592):
//
//	"Whenever this creature attacks, you may pay {W}. If you do, it gains
//	 flying until end of turn."
//
// Estwald Shieldbasher's attack trigger with an optional payment (MayPay,
// CR 118.12) and flying for indestructible. Gaining flying after blockers
// are declared changes nothing about the blocks (CR 702.9b), so the
// payment matters only before blocks, which is when the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02abe85a-c603-4963-aba4-524cc80847ea",
		Name:         "Order of the Golden Cricket",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Order of the Golden Cricket — you may pay {W}; if you do, it gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{W}",
						Question: "Order of the Golden Cricket — pay {W} to give it flying until end of turn?",
						Then:     orderOfTheGoldenCricketPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// orderOfTheGoldenCricketPaid grants the keyword. An on-disk key: never
// renamed, never reused.
var orderOfTheGoldenCricketPaid = PayUnlessBranches("pay-unless/order-of-the-golden-cricket-flying",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil || sourceIsNewObject(ctx.Game, ctx.Item) {
			return nil
		}
		return GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"},
			Label: "Order of the Golden Cricket — flying"}.Apply(ctx)
	}, nil)
