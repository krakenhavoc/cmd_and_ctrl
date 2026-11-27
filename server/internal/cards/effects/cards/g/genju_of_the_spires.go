package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genju of the Spires — Enchantment — Aura {R} (Betrayers of Kamigawa, S82 OLD b594):
//
//	"Enchant Mountain
//	 {2}: Enchanted Mountain becomes a 6/1 red Spirit creature until end of turn. It's
//	 still a land.
//	 When enchanted Mountain is put into a graveyard, you may return this card from your
//	 graveyard to your hand."
//
// Genju of the Falls' shape (GenjuAbility, GenjuReturnTrigger) with a 6/1 red body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c040882e-48ba-469e-9fc6-02e8b68dd58c",
		Name:         "Genju of the Spires",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.Subtype("Mountain")),
		Activated: []effects.ActivatedAbility{effects.GenjuAbility("Genju of the Spires",
			"{2}: Enchanted Mountain becomes a 6/1 red Spirit creature until end of turn. It's still a land.",
			effects.GenjuForm{Colors: []string{"R"}, Power: 6, Toughness: 1})},
		Triggered: []game.TriggeredAbility{effects.GenjuReturnTrigger("Genju of the Spires")},
	})
}
