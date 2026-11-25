package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bind the Monster — Enchantment — Aura {U} (Kaldheim, #5505):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature. It deals damage to
//	 you equal to its power.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Waterknot's lock and entry tap. The damage is dealt BY the enchanted
// creature (so lifelink, deathtouch and protection read it), to the
// Aura's controller, for the power it has as the trigger resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "8d1a3935-8a8c-4754-b474-189611188bc5",
		Name:                  "Bind the Monster",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bind the Monster — tap enchanted creature; it deals damage to you equal to its power", effects.TapEnchantedAndItDamagesYou),
		},
	})
}
