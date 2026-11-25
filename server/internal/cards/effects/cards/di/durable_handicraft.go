package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Durable Handicraft — Enchantment {1}{G} (Kaladesh, #5539):
//
//	"Whenever a creature you control enters, you may pay {1}. If you do, put a
//	 +1/+1 counter on that creature.
//	 {5}{G}, Sacrifice this enchantment: Put a +1/+1 counter on each creature
//	 you control."
//
// Tainted Observer's creature-entered trigger (here without "another", the
// Handicraft not being a creature) with its "you may pay. If you do" (MayPay)
// and Mirrorworks' read of the entering card off the trigger's event: the
// counter goes on that creature only if it is still on the battlefield. The
// activated ability sacrifices the Handicraft as a cost and puts a counter on
// each creature its controller controls as it resolves (Munda's Vanguard's
// group put).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b85a5d4-1cc3-412d-99e7-778af68d79a5",
		Name:         "Durable Handicraft",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
				"Durable Handicraft — you may pay {1}; if you do, put a +1/+1 counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Durable Handicraft — pay {1} to put a +1/+1 counter on that creature?",
						Then:     durableHandicraftPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}, Sacrifice this enchantment: Put a +1/+1 counter on each creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{5}{G}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B11PutCountersOnEachCreatureYouControl(g, item, 1)
			},
		}},
	})
}

// durableHandicraftPaid is the "if you do": a +1/+1 counter on the creature
// that entered, if it is still on the battlefield. An on-disk key.
var durableHandicraftPaid = effects.PayUnlessBranches("pay-unless/durable-handicraft-counter",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		id := ctx.Trigger().Event.CardID
		if !effects.OnBattlefield(ctx.Game, id) {
			return nil
		}
		return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}, nil)
