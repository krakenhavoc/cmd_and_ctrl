package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const curiousInquiryInvestigate = "curious-inquiry/investigate"

// Curious Inquiry — Enchantment — Aura {U} (Murders at Karlov Manor,
// #4469):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has 'Whenever this creature deals
//	 combat damage to a player, investigate.' (Create a Clue token. It's
//	 an artifact with '{2}, Sacrifice this token: Draw a card.')"
//
// Curious Obsession without the end-step sacrifice. The investigate is
// GRANTED to the creature (ADR 0093), so the creature's controller makes
// the Clue, and a creature that loses all abilities loses it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52fe8101-82af-4aed-a395-54e7b7639bd2",
		Name:         "Curious Inquiry",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: curiousInquiryInvestigate,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Curious Inquiry — investigate",
					effects.Do(effects.Investigate{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, investigate.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(curiousInquiryInvestigate),
		},
	})
}
