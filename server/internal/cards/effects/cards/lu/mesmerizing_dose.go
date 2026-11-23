package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mesmerizing Dose — Enchantment — Aura {1}{U}{U} (Phyrexia: All Will Be One,
// S82):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature, then proliferate.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Claustrophobia's shape (the tap reads the Aura's attachment, last-known if
// it has left, CR 608.2h) with the proliferate after it, so the tapped
// creature's counters are already as the tap left them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "195ab0c7-0848-4803-a299-9155e1f56a85",
		Name:                  "Mesmerizing Dose",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mesmerizing Dose — tap enchanted creature, then proliferate",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.TapEnchantedCreatureOnEntry(g, item); err != nil {
						return err
					}
					return effects.Proliferate{}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
