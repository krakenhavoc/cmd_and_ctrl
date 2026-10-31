package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Favored Hoplite — Creature — Human Soldier {W}, 1/2:
//
//	"Heroic — Whenever you cast a spell that targets this creature, put a
//	 +1/+1 counter on this creature and prevent all damage that would be
//	 dealt to it this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): heroic (CR 207.2c) is the cast
// trigger in shield_families_recipient.go, read off the spell's stack
// item as it is cast. The effect is the counter, then the not-one-use
// shield pinned to this creature. A Hoplite that has left and come back
// before the trigger resolves is a new object, and gets neither
// (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "145a15f9-89a9-4eb7-9924-e768f17e7e68",
		Name:         "Favored Hoplite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			// The shield waits for the counter's prompt (a CR 616 order; CR 608.2c, #4155).
			heroic("Favored Hoplite — a +1/+1 counter, and prevent all damage to it this turn (heroic)", Do(
				Instruction(plusOneCounterOnSourceIfSameObject),
				Instruction(func(ctx *Context) error {
					return shieldThis(ctx.Game, ctx.Item, PreventDamageFromSource{})
				}),
			)),
		},
	})
}
