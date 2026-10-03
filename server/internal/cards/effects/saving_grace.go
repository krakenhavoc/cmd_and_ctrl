package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saving Grace — Enchantment — Aura {1}{W}:
//
//	"Flash
//	 Enchant creature you control
//	 When this Aura enters, all damage that would be dealt this turn to
//	 you and permanents you control is dealt to enchanted creature
//	 instead.
//	 Enchanted creature gets +0/+3."
//
// ADR 0108 §9 (#1905): the trigger makes a redirection for the rest of
// the turn to the creature the Aura enchants as it resolves. The rulings:
// the redirection stays with that creature if the Aura leaves; it stops
// if the creature is gone or no longer a creature; damage dealt to the
// creature itself stays where it is.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "50ddbea3-7ef4-4f6a-83c9-0c3ea1dfa3c9",
		Name:         "Saving Grace",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Static:       []game.StaticAbility{PumpAttached(0, 3)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Saving Grace — damage to you and your permanents is dealt to enchanted creature instead",
				func(g *game.Game, item *game.StackItem) error {
					return RedirectDamage{Protect: ShieldYouAndPermanentsYouControl, To: RedirectToEnchanted}.Apply(NewContext(g, item))
				}),
		},
	})
}
