package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Qala, Ajani's Pridemate — Legendary Creature — Cat Warrior {3}{W}, 3/3
// (Foundations Jumpstart, #5929):
//
//	"Whenever Qala attacks, other attacking creatures you control get +X/+0
//	 until end of turn, where X is the number of counters on Qala.
//	 Whenever you gain life, put a +1/+1 counter on Qala.
//	 {3}{W}: You gain 1 life."
//
// The counter trigger is Ajani's Pridemate's (one per life-gain event, paying
// that event's LifeGainPayoff). The attack trigger counts counters of every
// kind on Qala as it resolves (CR 608.2h: last-known information if she has
// left), and boosts the other attacking creatures its controller controls then,
// evaluated once (CR 611.2c): a creature that joins combat later gets nothing.
// With no counters, X is 0 and the trigger changes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1fc1b0d7-7f15-4d00-bc8f-e06489afd425",
		Name:         "Qala, Ajani's Pridemate",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Qala — other attacking creatures you control get +X/+0 until end of turn, where X is the number of counters on Qala",
				effects.Do(effects.Instruction(qalaPumpAttackers))),
			effects.LifeGainPaying(
				effects.WheneverYouGainLife("Qala — put a +1/+1 counter on her", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				game.LifeGainPayoff{Counters: 1}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{W}"),
			Effect:  effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
}

func qalaPumpAttackers(ctx *effects.Context) error {
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	x := 0
	for _, n := range info.Counters {
		if n > 0 {
			x += n
		}
	}
	if x == 0 {
		return nil
	}
	return effects.BoostUntilEOT{
		Match: effects.And(effects.Creature(), effects.AttackingCreature(), effects.YouControl(), effects.OtherThan(ctx.Item.SourceCardID)),
		Power: x,
		Label: "Qala — other attackers get +X/+0",
	}.Apply(ctx)
}
