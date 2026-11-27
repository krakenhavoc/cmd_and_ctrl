package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genju of the Falls — Enchantment — Aura {U} (Betrayers of Kamigawa, S82 OLD b594):
//
//	"Enchant Island
//	 {2}: Enchanted Island becomes a 3/2 blue Spirit creature with flying until end of
//	 turn. It's still a land.
//	 When enchanted Island is put into a graveyard, you may return this card from your
//	 graveyard to your hand."
//
// The genju cycle (GenjuAbility, GenjuReturnTrigger): the land keeps its types and gains
// the creature ones, so a land that has been under your control since your turn began can
// attack and tap for mana only when it is no longer summoning-sick (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c546747a-1935-4c59-a45f-cad7da41fe38",
		Name:         "Genju of the Falls",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.Subtype("Island")),
		Activated: []effects.ActivatedAbility{effects.GenjuAbility("Genju of the Falls",
			"{2}: Enchanted Island becomes a 3/2 blue Spirit creature with flying until end of turn. It's still a land.",
			effects.GenjuForm{Colors: []string{"U"}, Power: 3, Toughness: 2, Keywords: []string{"flying"}})},
		Triggered: []game.TriggeredAbility{effects.GenjuReturnTrigger("Genju of the Falls")},
	})
}
