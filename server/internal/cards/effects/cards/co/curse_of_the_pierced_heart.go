package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of the Pierced Heart — Enchantment — Aura Curse {1}{R}:
//
//	"Enchant player
//	 At the beginning of enchanted player's upkeep, this Aura deals 1 damage
//	 to that player or a planeswalker that player controls."
//
// The Aura's controller picks the recipient as the trigger resolves; it is not
// a target. The Aura is the damage's source (an Aura deals no combat damage and
// has no damage-linked abilities of its own).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f22c3624-6bd5-4761-b8ad-ac301ad9351c",
		Name:         "Curse of the Pierced Heart",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.OldB444EnchantedPlayersUpkeepTrigger("Curse of the Pierced Heart — 1 damage to that player or a planeswalker that player controls",
				effects.OldB444DamageEnchantedPlayerOrAPlaneswalker),
		},
	})
}
