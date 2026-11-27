package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genju of the Cedars — Enchantment — Aura {G} (Betrayers of Kamigawa, S82 OLD b594):
//
//	"Enchant Forest
//	 {2}: Enchanted Forest becomes a 4/4 green Spirit creature until end of turn. It's
//	 still a land.
//	 When enchanted Forest is put into a graveyard, you may return this card from your
//	 graveyard to your hand."
//
// Genju of the Falls' shape (GenjuAbility, GenjuReturnTrigger) with a 4/4 green body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ada3c285-3449-4037-8b44-42e128456a65",
		Name:         "Genju of the Cedars",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.Subtype("Forest")),
		Activated: []effects.ActivatedAbility{effects.GenjuAbility("Genju of the Cedars",
			"{2}: Enchanted Forest becomes a 4/4 green Spirit creature until end of turn. It's still a land.",
			effects.GenjuForm{Colors: []string{"G"}, Power: 4, Toughness: 4})},
		Triggered: []game.TriggeredAbility{effects.GenjuReturnTrigger("Genju of the Cedars")},
	})
}
