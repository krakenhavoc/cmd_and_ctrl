package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eidolon of Obstruction — Enchantment Creature — Spirit {1}{W}, 2/1 (Theros
// Beyond Death, #5514):
//
//	"First strike
//	 Loyalty abilities of planeswalkers your opponents control cost {1} more to
//	 activate."
//
// An activation-scoped cost increase (CR 601.2f, CR 606.4: a loyalty ability's cost
// gains the {1} beside its loyalty counter), read through the cost query's Loyalty
// bit and the ability's source being controlled by an opponent of the Eidolon's
// controller. It stacks with a second Eidolon.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15d38195-b661-42c1-82c9-1af2a3ba2ea5",
		Name:            "Eidolon of Obstruction",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		CostModifiers: []game.CostModifier{
			effects.ActivationCostsMore(1, "Loyalty abilities of planeswalkers your opponents control cost {1} more to activate.",
				effects.LoyaltyAbilityOfAnOpponentsPlaneswalker()),
		},
	})
}
