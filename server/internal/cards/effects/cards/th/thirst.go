package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thirst — Enchantment — Aura {2}{U} (Mirage, S82 OLD batch 811, #5633):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 At the beginning of your upkeep, sacrifice this Aura unless you pay {U}."
//
// Bitter Chill's tap and untap lock, with Spindrift Drake's upkeep
// pay-or-sacrifice on the Aura (the controller is asked as it resolves, and the
// table waits in the upkeep).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "61b3ad9d-485f-4479-b18f-506847af43f2",
		Name:                  "Thirst",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thirst — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
			effects.AtYourUpkeep("Thirst — sacrifice this Aura unless you pay {U}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{U}",
						Question: "Thirst — pay {U} or sacrifice Thirst?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
