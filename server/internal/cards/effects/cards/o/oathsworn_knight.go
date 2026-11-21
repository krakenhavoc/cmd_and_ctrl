package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oathsworn Knight — Creature — Human Knight {1}{B}{B}, 0/0:
//
//	"This creature enters with four +1/+1 counters on it.
//	 This creature attacks each combat if able.
//	 If damage would be dealt to this creature while it has a +1/+1 counter on it, prevent that damage and remove a +1/+1 counter from it."
//
// ADR 0108 §8 (#1906): a prevention static with an additional effect, one
// application per recipient in a damage instance — several creatures
// blocking it at once cost one counter, not one per point or per source
// — and damage that can't be prevented still removes one (CR 615.12; the
// rulings). It applies only while the Knight has a +1/+1 counter.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16c293c8-bd1d-4db9-b197-da14f2f37cb1",
		Name:         "Oathsworn Knight",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 4, "Oathsworn Knight: enters with four +1/+1 counters"),
			effects.PreventDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToThisCreature,
				While: effects.WhileItHasAPlusOneCounter,
				Then:  effects.RemoveACounterFromThisBody,
				Label: "Oathsworn Knight — prevent damage to it and remove a +1/+1 counter",
			}),
		},
	})
}
