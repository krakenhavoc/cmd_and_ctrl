package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cruel Sadist — Creature — Human Assassin {B}, 1/1 (Jumpstart 2022, #5948):
//
//	"{B}, {T}, Pay 1 life: Put a +1/+1 counter on this creature.
//	 {2}{B}, {T}, Remove X +1/+1 counters from this creature: It deals X
//	 damage to target creature."
//
// The life and the counters are costs, paid as the ability is activated
// (CR 602.2b, 119.4): a player at 0 life cannot pay, and X is announced with
// the cost, from 0 up to the +1/+1 counters the Sadist has, so a response
// cannot spend the same counter twice. The damage is the Sadist's, read
// from its last known information if it has gone (CR 608.2h); X is the
// number of counters removed as the cost was paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fceca3c-fc10-4e46-a727-122aafcef349",
		Name:         "Cruel Sadist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}, {T}, Pay 1 life: Put a +1/+1 counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerValue}, // ADR 0142: a counter on itself answers nothing on the stack
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost(), effects.PayLife(1)),
				Effect:  effects.PlusOneCountersOnThis(1),
			},
			{
				Label:   "{2}{B}, {T}, Remove X +1/+1 counters from this creature: It deals X damage to target creature.",
				Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
				Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost(), effects.RemoveCountersXFromThis(game.CounterPlusOne, 0)),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := ctx.CountersRemoved()
					if x <= 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}.Apply(ctx)
						}
					}
					return nil
				},
			},
		},
	})
}
