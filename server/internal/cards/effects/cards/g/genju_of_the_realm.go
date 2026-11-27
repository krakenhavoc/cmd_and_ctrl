package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genju of the Realm — Legendary Enchantment — Aura {W}{U}{B}{R}{G} (Betrayers of
// Kamigawa, S82 OLD b594):
//
//	"Enchant land
//	 {2}: Enchanted land becomes a legendary 8/12 Spirit creature with trample until end
//	 of turn. It's still a land.
//	 When enchanted land is put into a graveyard, you may return Genju of the Realm from
//	 your graveyard to your hand."
//
// Genju of the Falls' shape (GenjuAbility, GenjuReturnTrigger) with no colour change: the
// land keeps its own colour (none), gains Legendary, Creature and Spirit, and is 8/12 with
// trample.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1da3a0c0-43f0-4feb-b20a-414b023e710c",
		Name:         "Genju of the Realm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Activated: []effects.ActivatedAbility{effects.GenjuAbility("Genju of the Realm",
			"{2}: Enchanted land becomes a legendary 8/12 Spirit creature with trample until end of turn. It's still a land.",
			effects.GenjuForm{Power: 8, Toughness: 12, Keywords: []string{"trample"}, Legendary: true})},
		Triggered: []game.TriggeredAbility{effects.GenjuReturnTrigger("Genju of the Realm")},
	})
}
