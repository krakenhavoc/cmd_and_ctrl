package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slow Motion — Enchantment — Aura {2}{U}:
//
//	"Enchant creature
//	 At the beginning of the upkeep of enchanted creature's controller, that
//	 player sacrifices that creature unless they pay {2}.
//	 When this Aura is put into a graveyard from the battlefield, return it to
//	 its owner's hand."
//
// The upkeep is the enchanted creature's controller's, who is asked to pay;
// declining sacrifices the creature the Aura enchants as the trigger resolves.
// The second ability triggers when the Aura goes from the battlefield to a
// graveyard, whether the creature died, it was sacrificed or the Aura was
// destroyed.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "151388a4-a24b-4089-a323-e7c1d5e6490a",
		Name:         "Slow Motion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Slow Motion — that player sacrifices that creature unless they pay {2}",
				effects.OldB719SacrificeTheEnchantedCreatureUnlessPay("Slow Motion", "{2}")),
			effects.WhenThisDies("Slow Motion — return it to its owner's hand", effects.ReturnThisCardFromYourGraveyardToHand),
		},
	})
}
