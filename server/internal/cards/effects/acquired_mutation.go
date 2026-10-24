package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acquired Mutation — Enchantment — Aura {2}{R}:
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and is goaded. (It attacks each
//	 combat if able and attacks a player other than you if able.)
//	 Whenever enchanted creature attacks, defending player gets two
//	 rad counters."
//
// #2809. Shiny Impetus's pair (PumpAttached, and GoadAttached, the goad
// as a requirement pair that ends when the Aura falls off) with a
// rad-counter trigger. The trigger is the Aura's, so the Aura's
// controller is the one who gives the counters, whoever
// controls the creature.
//
// "Defending player" is the player behind the attack's target
// (CR 508.5): the player attacked, or the controller of the attacked
// planeswalker or battle. It is read off the attack event, so it is
// still right when the creature has left combat by the time the
// trigger resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9ecd53b0-4551-4505-b5d2-02ae49d4cbdb",
		Name:         "Acquired Mutation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GoadAttached(),
		},
		Triggered: []game.TriggeredAbility{
			WheneverEnchantedCreatureAttacks("Acquired Mutation — defending player gets two rad counters",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					defender := g.DefendingPlayerForAttackForEffect(item.Trigger.Event.Target)
					return playerGetsRadCounters(g, item.Controller, defender, 2)
				}),
		},
	})
}
