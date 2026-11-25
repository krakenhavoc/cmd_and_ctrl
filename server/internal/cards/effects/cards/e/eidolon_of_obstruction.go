package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eidolon of Obstruction — Enchantment Creature — Spirit {1}{W}, 2/1
// (Theros Beyond Death, #6148):
//
//	"First strike
//	 Loyalty abilities of planeswalkers your opponents control cost {1}
//	 more to activate."
//
// An activation cost increase (CR 602.2b, 601.2f) scoped to loyalty
// abilities (AbilityCostSubject.Loyalty, CR 606.1) of planeswalkers an
// opponent of Eidolon's controller controls: their loyalty cost is
// unchanged and {1} is added beside it (CR 606.4: "This cost may be
// modified by other effects"), so without the mana the ability can't be
// activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15d38195-b661-42c1-82c9-1af2a3ba2ea5",
		Name:            "Eidolon of Obstruction",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		CostModifiers: []game.CostModifier{effects.ActivationCostsMore(1,
			"Loyalty abilities of planeswalkers your opponents control cost {1} more to activate.",
			eidolonOfObstructionLoyaltyAbilityOfAnOpponentsPlaneswalker)},
	})
}

// eidolonOfObstructionLoyaltyAbilityOfAnOpponentsPlaneswalker passes for a
// loyalty ability of a planeswalker controlled by an opponent of the
// modifier's controller.
func eidolonOfObstructionLoyaltyAbilityOfAnOpponentsPlaneswalker(q game.CostQuery) bool {
	return q.Ability != nil && q.Ability.Loyalty && !q.Ability.Mana && q.Card.IsPlaneswalker() &&
		q.Card.Controller != q.Source.Controller
}
