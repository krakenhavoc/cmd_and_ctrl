package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seymour Flux — Legendary Creature — Spirit Avatar {4}{B}, 5/5 (Final
// Fantasy, #4462):
//
//	"At the beginning of your upkeep, you may pay 1 life. If you do, draw
//	 a card and put a +1/+1 counter on Seymour Flux."
//
// "You may pay 1 life. If you do" is a payment, not a yes/no: MayPay
// offers LifePayment(1) as the trigger resolves, a player who can't pay
// life (CR 119.4) is not asked, and the paid branch runs only once the
// life is really paid. The counter goes on Seymour only while he is still
// the permanent the trigger came from (CR 400.7); the card is drawn
// either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7312efc2-8a02-47cd-874b-c3338241435e",
		Name:         "Seymour Flux",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Seymour Flux — you may pay 1 life to draw a card and grow",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Pay:      LifePayment(1),
						Question: "Seymour Flux — pay 1 life to draw a card and put a +1/+1 counter on Seymour Flux?",
						Then:     seymourFluxPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// seymourFluxPaid is "draw a card and put a +1/+1 counter on Seymour
// Flux". An on-disk key: never renamed, never reused.
var seymourFluxPaid = PayUnlessBranches("pay-unless/seymour-flux-draw-and-counter",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
			return err
		}
		if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}, nil)
