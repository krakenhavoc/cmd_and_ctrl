package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:              "48533a2d-1bdb-40a6-939e-9c8337e0dc5a",
		Name:                  "Sleep with the Fishes",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{WhenThisEnters("Sleep with the Fishes — tap enchanted creature and create a Fish",
			func(g *game.Game, item *game.StackItem) error {
				if err := tapEnchantedCreatureOnEntry(g, item); err != nil {
					return err
				}
				return CreateToken{Controller: item.Controller, Template: S82SncB07FishToken(), N: 1}.Apply(NewContext(g, item))
			})},
	})
}
