package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jack of Hearts, Volatile Hero — {1}{R}{R} Legendary Creature — Alien
// Human Hero 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flying, haste
//	 When Jack of Hearts dies, he deals damage equal to his power to each
//	 creature.
//	 Power-up — {4}{R}{R}: Discard your hand, then draw three cards. Put
//	 two +1/+1 counters on Jack of Hearts. (Activate each power-up ability
//	 only once. Reduce the cost by his mana cost if he entered this turn.)"
//
// The death trigger reads his last-known power (CR 608.2h, the departure
// record through ctx.TriggeringPermanent, clamped at zero) and deals it to
// every creature, himself being already gone. It is a creature wipe, so the
// trigger declares its sweep. Power-up is the engine's (CR 702.193): the
// hand goes first, the draw waits for any prompt the discard asks, and the
// counters land on him if he is still on the battlefield.
//
// No simplification.
func init() {
	ab := effects.PowerUp(
		"{4}{R}{R}: Discard your hand, then draw three cards. Put two +1/+1 counters on Jack of Hearts.",
		effects.ManaCost("{4}{R}{R}"),
		effects.Do(effects.Instruction(func(ctx *effects.Context) error {
			_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
			return err
		}), effects.DrawCards{N: 3}, effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}))
	ab.Purpose.Answers = game.AnswerPump
	effects.Register(effects.Spec{
		OracleID:        "25b30117-03ef-49a5-8743-2d1c167d4bc3",
		Name:            "Jack of Hearts, Volatile Hero",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisDies("Jack of Hearts — he deals damage equal to his power to each creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, _ := ctx.TriggeringPermanent()
					return effects.DamageEachMatching(ctx, effects.Creature(), max(info.Power, 0))
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}}),
		},
		Activated: []effects.ActivatedAbility{ab},
	})
}
