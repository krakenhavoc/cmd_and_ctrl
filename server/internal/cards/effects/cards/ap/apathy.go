package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Apathy — Enchantment — Aura {U} (Weatherlight):
//
//	"Enchant creature
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 At the beginning of the upkeep of enchanted creature's controller, that
//	 player may discard a card at random. If the player does, untap that
//	 creature."
//
// Dance of the Dead's shape with a random discard in place of mana: the upkeep
// is the enchanted creature's controller's, who is the one offered the discard
// (and needs a card in hand to say yes, CR 118.3). The untap is the creature the
// Aura enchants as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "09ac5adf-f722-417e-b0fc-7ed0d8c9abd6",
		Name:                  "Apathy",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Apathy — that player may discard a card at random to untap that creature",
				effects.OldB774MayDiscardAtRandomToUntapEnchantedCreature("Apathy")),
		},
	})
}
