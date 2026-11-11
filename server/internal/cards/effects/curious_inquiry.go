package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "52fe8101-82af-4aed-a395-54e7b7639bd2",
		Name:         "Curious Inquiry",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: curiousInquiryInvestigate,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Curious Inquiry — investigate",
					Do(Investigate{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, investigate.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(curiousInquiryInvestigate),
		},
	})
}
