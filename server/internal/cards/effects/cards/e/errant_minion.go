package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Errant Minion — Enchantment — Aura {2}{U} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 At the beginning of the upkeep of enchanted creature's controller, that
//	 player may pay any amount of mana. This Aura deals 2 damage to that player.
//	 Prevent X of that damage, where X is the amount of mana that player paid
//	 this way."
//
// Karn, Living Legacy's "pay any amount of mana" (MayPayX); the payer chooses X
// as the trigger resolves, then the Aura deals 2 damage less X (never below 0).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "80459049-bd89-4736-82f8-03c603644033",
		Name:         "Errant Minion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Errant Minion — that player may pay any amount of mana; this Aura deals 2 damage to them, X prevented",
				effects.ErrantMinionPaymentB829),
		},
	})
}
