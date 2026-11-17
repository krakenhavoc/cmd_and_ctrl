package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unstable Mutation — Enchantment — Aura {U} (Arabian Nights, #5641):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3.
//	 At the beginning of the upkeep of enchanted creature's controller,
//	 put a -1/-1 counter on that creature."
//
// The trigger is Super Intelligence's When (the upkeep of whoever
// controls the enchanted creature now, which need not be the Aura's
// controller), and the counter is Glistening Oil's body
// (glisteningOilCounter): it goes on the
// creature the Aura enchants as the trigger resolves. The counters are
// on the creature, so they stay when the Aura leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "278b237e-9699-43eb-a03e-0b68eccc08b3",
		Name:         "Unstable Mutation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, UpkeepOfEnchantedCreaturesController,
				"Unstable Mutation — put a -1/-1 counter on that creature",
				glisteningOilCounter),
		},
	})
}
