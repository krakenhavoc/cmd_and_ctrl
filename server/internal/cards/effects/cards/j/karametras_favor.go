package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const karametrasFavorGrant = "karametras-favor/any-color"

// Karametra's Favor — Enchantment — Aura {1}{G} (Starter Commander Decks):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature has "{T}: Add one mana of any color.""
//
// Rousing Read's enters draw (a trigger on the stack) and Paradise Mantle's
// granted mana ability, which follows the Aura and is the enchanted
// creature's own, so CR 302.6 still applies to it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc3e48c9-e303-435c-bdb1-7dc7881a9858",
		Name:         "Karametra's Favor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Grants:       []effects.AbilityGrant{effects.AnyColorManaGrant(karametrasFavorGrant)},
		Static:       []game.StaticAbility{effects.GrantAbilitiesToAttached(karametrasFavorGrant)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Karametra's Favor — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
