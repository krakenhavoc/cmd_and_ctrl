package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wildborn Preserver — Creature — Elf Archer {1}{G}, 2/2 (Foundations,
// #4465):
//
//	"Flash
//	 Reach
//	 Whenever another non-Human creature you control enters, you may pay
//	 {X}. When you do, put X +1/+1 counters on this creature."
//
// Tester of the Tangential's "you may pay {X}. When you do" (MayPayX): X is
// chosen as the trigger resolves (CR 608.2d) and paid through the ordinary
// "you may pay" prompt, and paying nothing is declining. Paying makes a
// CR 603.12 reflexive trigger of its own that carries X, so players may act
// between the payment and the counters (the ruling). "Non-Human" reads the
// entering creature's types through the layered view, so a changeling is a
// Human. The counters go on this creature only if it is still the same
// object on the battlefield (CR 400.7).
//
// No simplification.
func init() {
	const label = "Wildborn Preserver — you may pay {X}; put X +1/+1 counters on this creature"
	Register(Spec{
		OracleID:        "55cec4bc-3de6-4c0d-a4cc-c5a30849fbda",
		Name:            "Wildborn Preserver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && !c.HasSubtypeWord("Human")
			}, label, func(g *game.Game, item *game.StackItem) error {
				return MayPayX{
					Cost:  "{X}",
					Label: "Wildborn Preserver",
					Buys:  "put X +1/+1 counters on this creature",
					Unit:  game.PayAmountCounters,
					Then:  wildbornPreserverPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// wildbornPreserverPaid is the paid half, keyed: X rides the prompt.
var wildbornPreserverPaid = PayUnlessBranches("pay-unless/wildborn-preserver-counters",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Wildborn Preserver — put X +1/+1 counters on this creature",
			Body:   wildbornPreserverCountersBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

// wildbornPreserverCountersBody is the reflexive trigger: "put X +1/+1
// counters on this creature".
var wildbornPreserverCountersBody = game.ReflexiveBody("wildborn-preserver/counters", wildbornPreserverCounters, nil)

func wildbornPreserverCounters(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 || !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: p.Amount}.Apply(NewContext(g, item))
}
