package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ignoble Soldier — Creature — Human Soldier {2}{W}, 3/1:
//
//	"Whenever this creature becomes blocked, prevent all combat damage that would be dealt by it this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the trigger is the
// EventBecomesBlocked the block declaration's lock-in emits once per
// blocked attacker (CR 509.1h, #830), and the shield's source is the
// Soldier itself while it is still the same object (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "52a61983-53e2-440f-af03-cebe57b102f2",
		Name:         "Ignoble Soldier",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesBlocked, Self, "Ignoble Soldier — prevent all combat damage it would deal this turn",
				func(g *game.Game, item *game.StackItem) error {
					return shieldAgainstThisCombatDamage(NewContext(g, item))
				}),
		},
	})
}
