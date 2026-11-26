package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nemesis of Mortals — Creature — Snake {4}{G}{G}, 5/5 (Theros):
//
//	"This spell costs {1} less to cast for each creature card in your
//	 graveyard.
//	 {7}{G}{G}: Monstrosity 5. This ability costs {1} less to activate for each
//	 creature card in your graveyard. (If this creature isn't monstrous, put
//	 five +1/+1 counters on it and it becomes monstrous.)"
//
// Molderhulk's spell-cost reduction and Battlefield Butcher's ability-cost
// reduction, both counted from the creature cards in the controller's graveyard
// when the cost is determined (CR 601.2f, 602.2b); only the generic part is
// reduced.
//
// No simplification.
func init() {
	monstrosity := effects.Monstrosity(effects.ManaCost("{7}{G}{G}"), 5)
	monstrosity.Label = "{7}{G}{G}: Monstrosity 5. This ability costs {1} less to activate for each creature card in your graveyard."
	monstrosity.CostModifiers = []game.CostModifier{
		effects.CostsLessEach(effects.CreatureCardsInYourGraveyard,
			"This ability costs {1} less to activate for each creature card in your graveyard."),
	}
	effects.Register(effects.Spec{
		OracleID:     "599fc217-e0a3-45ee-b9aa-cd7f181dcdf3",
		Name:         "Nemesis of Mortals",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.CreatureCardsInYourGraveyard,
				"This spell costs {1} less to cast for each creature card in your graveyard."),
		},
		Activated: []effects.ActivatedAbility{monstrosity},
	})
}
