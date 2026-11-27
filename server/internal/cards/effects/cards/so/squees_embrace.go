package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squee's Embrace — Enchantment — Aura {R}{W} (Apocalypse, S82 OLD, #5614):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 When enchanted creature dies, return that card to its owner's hand."
//
// Demonic Vigor's shape at +2/+2: a layer 7c pump on the host, and a dies
// trigger that brings the card back from the graveyard it just reached. A token
// has no card to return, and a card that has left the graveyard by resolution
// stays where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "095f8dac-15b8-4c28-ae33-6dc71152bcc7",
		Name:         "Squee's Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Squee's Embrace — return that card to its owner's hand", effects.S82OldB212ReturnDiedCardToOwnersHand),
		},
	})
}
