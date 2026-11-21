package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sleep with the Fishes — Enchantment — Aura {2}{U}{U} (Streets of New
// Capenna, #5491):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature and you create a 1/1
//	 blue Fish creature token with "This token can't be blocked."
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Charmed Sleep's shape with a Fish for the Aura's controller: the entry
// trigger reads the attachment as it resolves (CR 608.2h), the tap and the
// token are one ability, and the untap-step hold lasts only while the Aura
// enchants the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "48533a2d-1bdb-40a6-939e-9c8337e0dc5a",
		Name:                  "Sleep with the Fishes",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{effects.WhenThisEnters("Sleep with the Fishes — tap enchanted creature and create a Fish",
			func(g *game.Game, item *game.StackItem) error {
				if err := effects.TapEnchantedCreatureOnEntry(g, item); err != nil {
					return err
				}
				return effects.CreateToken{Controller: item.Controller, Template: effects.S82SncB07FishToken(), N: 1}.Apply(effects.NewContext(g, item))
			})},
	})
}
