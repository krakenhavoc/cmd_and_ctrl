package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oath of Ajani — {G}{W} Legendary Enchantment (Aether Revolt, #5536):
//
//	"When Oath of Ajani enters, put a +1/+1 counter on each creature you
//	 control.
//	 Planeswalker spells you cast cost {1} less to cast."
//
// The discount is the cost modifier Tam, the Possibility uses, over the
// controller's own planeswalker spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ec39c8f-eb1e-4881-a749-0c188788b612",
		Name:         "Oath of Ajani",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Oath of Ajani — put a +1/+1 counter on each creature you control", effects.B11PutCounterOnEachCreatureYouControl),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Planeswalker spells you cast cost {1} less to cast.",
				effects.YourSpell(), func(q game.CostQuery) bool { return q.Card.IsPlaneswalker() }),
		},
	})
}
