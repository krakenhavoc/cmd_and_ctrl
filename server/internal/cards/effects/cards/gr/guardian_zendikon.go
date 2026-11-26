package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian Zendikon — Enchantment — Aura {2}{W} (Conspiracy, S82 OLD batch
// 362, #5557):
//
//	"Enchant land
//	 Enchanted land is a 2/6 white Wall creature with defender. It's still a
//	 land.
//	 When enchanted land dies, return that card to its owner's hand."
//
// The Zendikon animation (S82OldB100AnimateAttachedLand: Creature and the
// Wall subtype in layer 4, white in layer 5, 2/6 in layer 7b) with defender
// granted in layer 6, all kept while the Aura is on it. The dies trigger
// returns the card to its owner's hand.
//
// No simplification.
func init() {
	grant := effects.GrantToAttached("defender")
	grant.ContinuesAfterRemoval = true
	effects.Register(effects.Spec{
		OracleID:     "44154004-e50e-45d1-bde0-ba6e3354067c",
		Name:         "Guardian Zendikon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       append(effects.S82OldB100AnimateAttachedLand("W", "Wall", 2, 6), grant),
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Guardian Zendikon — return that card to its owner's hand",
				effects.S82OldB100ReturnTheDiedCardToHand),
		},
	})
}
