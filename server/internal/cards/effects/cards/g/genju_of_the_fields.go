package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genju of the Fields — Enchantment — Aura {W} (Betrayers of Kamigawa, S82 OLD b594):
//
//	"Enchant Plains
//	 {2}: Until end of turn, enchanted Plains becomes a 2/5 white Spirit creature with
//	 "Whenever this creature deals damage, its controller gains that much life." It's
//	 still a land.
//	 When enchanted Plains is put into a graveyard, you may return this card from your
//	 graveyard to your hand."
//
// Genju of the Falls' shape with a granted triggered ability (a bundle in Spec.Grants,
// given by the effect for the turn, ADR 0093): the life is the controller of the land's,
// and one trigger per damage dealt, whatever it is dealt to.
//
// No simplification.
const genjuOfTheFieldsGrant = "genju-of-the-fields/gain-life"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cd35d18-115b-4d7c-8d05-3d4c8303f46e",
		Name:         "Genju of the Fields",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.Subtype("Plains")),
		Grants: []effects.AbilityGrant{{
			Key:       genjuOfTheFieldsGrant,
			Triggered: []game.TriggeredAbility{effects.GapB139WheneverThisDealsDamageYouGainThatMuchLife("Genju of the Fields")},
			Text:      "Whenever this creature deals damage, its controller gains that much life.",
		}},
		Activated: []effects.ActivatedAbility{effects.GenjuAbility("Genju of the Fields",
			"{2}: Until end of turn, enchanted Plains becomes a 2/5 white Spirit creature with \"Whenever this creature deals damage, its controller gains that much life.\" It's still a land.",
			effects.GenjuForm{Colors: []string{"W"}, Power: 2, Toughness: 5, GrantKeys: []string{genjuOfTheFieldsGrant}})},
		Triggered: []game.TriggeredAbility{effects.GenjuReturnTrigger("Genju of the Fields")},
	})
}
