package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undergrowth Champion — Creature — Elemental {1}{G}{G}, 2/2:
//
//	"If damage would be dealt to this creature while it has a +1/+1 counter on it, prevent that damage and remove a +1/+1 counter from this creature.
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter on this creature."
//
// ADR 0108 §8 (#1906): a prevention static with an additional effect, one
// application per recipient in a damage instance — damage from several
// sources at once costs one counter — and damage that can't be prevented
// still removes one (CR 615.12; the rulings).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "cb0eb84d-b41b-4147-aa7c-089b7fabd835",
		Name:         "Undergrowth Champion",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				While: WhileItHasAPlusOneCounter,
				Then:  removeACounterFromThisBody,
				Label: "Undergrowth Champion — prevent damage to it and remove a +1/+1 counter",
			}),
		},
		Triggered: []game.TriggeredAbility{
			Landfall("Undergrowth Champion — put a +1/+1 counter on it", putCounterOnSelf),
		},
	})
}
