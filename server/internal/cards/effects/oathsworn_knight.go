package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "16c293c8-bd1d-4db9-b197-da14f2f37cb1",
		Name:         "Oathsworn Knight",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 4, "Oathsworn Knight: enters with four +1/+1 counters"),
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				While: WhileItHasAPlusOneCounter,
				Then:  removeACounterFromThisBody,
				Label: "Oathsworn Knight — prevent damage to it and remove a +1/+1 counter",
			}),
		},
	})
}
