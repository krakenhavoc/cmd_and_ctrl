package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anointer of Valor — Creature — Angel {5}{W}, 3/5 (Jumpstart 2022, #5948):
//
//	"Flying
//	 Whenever a creature attacks, you may pay {3}. When you do, put a
//	 +1/+1 counter on that creature."
//
// The attack trigger watches every attacking creature, anyone's. The
// "may pay {3}" is the parent ability's prompt (MayPay); once the mana is
// paid, the "when you do" is a CR 603.12 reflexive trigger carrying the
// attacker as its payload, so the counter goes on that creature as the
// reflexive ability resolves (CR 603.12a) and nothing happens if it has
// left the battlefield by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4f880f5-0639-4acb-a099-e48f5707f0db",
		Name:            "Anointer of Valor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventAttack {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			}, anointerOfValorLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				attacker := ctx.Trigger().Event.CardID
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{3}",
					Question: anointerOfValorLabel + "?",
					Then:     anointerOfValorPaid,
					Carry:    []uuid.UUID{attacker},
				}.Apply(ctx)
			}),
		},
	})
}

const anointerOfValorLabel = "Anointer of Valor — pay {3}; when you do, put a +1/+1 counter on that attacking creature"

// anointerOfValorPaid is the paid branch: the reflexive "when you do". An
// on-disk key.
var anointerOfValorPaid = effects.PayUnlessBranches("pay-unless/anointer-of-valor-counter",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Anointer of Valor — put a +1/+1 counter on that creature",
			Cards: r.Carry,
			Body:  anointerOfValorCounterBody,
		}.Apply(ctx)
	}, nil)

// anointerOfValorCounterBody puts the counter on the attacker the parent
// carried. An on-disk key.
var anointerOfValorCounterBody = game.ReflexiveBody("anointer-of-valor/counter", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}), nil)
