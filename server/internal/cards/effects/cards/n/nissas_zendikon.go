package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nissa's Zendikon — Enchantment — Aura {3}{G} (Zendikar Rising, #5508):
//
//	"Enchant land
//	 Enchanted land is a 4/4 Elemental creature with reach and haste. It's
//	 still a land.
//	 When enchanted land dies, return that card to its owner's hand."
//
// The Zendikon animation (S82OldB097AnimateAttachedLand): layer 4 type and
// subtype, layer 6 keywords, layer 7b 4/4, all kept while the Aura is on
// it. The dies trigger is read off the enchanted land as it goes to the
// graveyard and returns that card to its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b8c4cbe-b2fc-4d52-9382-1e03c40003be",
		Name:         "Nissa's Zendikon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       effects.S82OldB097AnimateAttachedLand("Elemental", 4, 4, "reach", "haste"),
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Nissa's Zendikon — return that card to its owner's hand",
				effects.S82OldB097ReturnTheDiedCardToHand),
		},
	})
}
