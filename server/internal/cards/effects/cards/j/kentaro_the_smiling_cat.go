package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kentaro, the Smiling Cat — Legendary Creature — Human Samurai {1}{W}, 2/1
// (Betrayers of Kamigawa, #5602):
//
//	"Bushido 1
//	 You may pay {X} rather than pay the mana cost for Samurai spells you
//	 cast, where X is that spell's mana value."
//
// Bushido 1 is the shared keyword trigger (CR 702.45a). The static grants every
// Samurai spell its controller casts an alternative cost (CR 118.9) of that
// many generic mana (GrantedAlternativeCost.PricedAtManaValue), derived from
// the spell's own mana value as the cost is chosen: coloured symbols are
// not owed, {X} in the printed cost counts as 0 (CR 107.3b), and a Samurai
// card with no mana cost costs {0}. Additional costs are still paid on top
// (CR 118.9d), and a commander's tax and cost reductions apply after the swap.
// The alternative cost is the player's choice, so the printed cost stays
// available.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfaf20a7-a11c-4f57-8eb6-b03b5223704d",
		Name:         "Kentaro, the Smiling Cat",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Bushido(1, "Kentaro, the Smiling Cat")},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{{
			Offer: game.AlternativeCost{
				Key:   effects.GrantedAltCostXForSamuraiSpells,
				Label: "Pay {X} rather than pay this spell's mana cost, where X is its mana value",
			},
			Spells:            game.PermissionFilter{SpellSubtype: "Samurai"},
			SpellsText:        "Samurai spells",
			PricedAtManaValue: true,
		}},
	})
}
