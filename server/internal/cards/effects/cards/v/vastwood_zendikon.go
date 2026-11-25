package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vastwood Zendikon — Enchantment — Aura {4}{G} (Jumpstart, #5510):
//
//	"Enchant land
//	 Enchanted land is a 6/4 green Elemental creature. It's still a land.
//	 When enchanted land dies, return that card to its owner's hand."
//
// The Zendikon animation: layer 4 creature type and subtype, layer 5 green,
// layer 7b 6/4, all kept while the Aura is on it. The dies trigger is read
// off the enchanted land as it goes to the graveyard and returns that card
// to its owner's hand (a card that has left the graveyard stays where it
// is).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10172c40-7169-4098-9452-b35797d27065",
		Name:         "Vastwood Zendikon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       effects.S82OldB100AnimateAttachedLand("G", "Elemental", 6, 4),
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Vastwood Zendikon — return that card to its owner's hand",
				effects.S82OldB100ReturnTheDiedCardToHand),
		},
	})
}
