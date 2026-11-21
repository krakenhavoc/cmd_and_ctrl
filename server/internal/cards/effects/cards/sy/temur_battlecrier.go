package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temur Battlecrier — Creature — Orc Ranger {G}{U}{R}, 4/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"During your turn, spells you cast cost {1} less to cast for each
//	 creature you control with power 4 or greater."
//
// Animar's scaling cost reduction (CostsLessEach) over every spell the
// controller casts, gated to their own turn (the cast's controller is the
// active player). The count is taken as the cost is determined (CR
// 601.2f): the creatures the controller has then with power 4 or greater,
// the Battlecrier itself included, read with layers applied. The discount
// spends generic mana only and stops at zero.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fadf984c-c224-432a-87b9-b34f929b8409",
		Name:         "Temur Battlecrier",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.StdB9BigCreaturesYouControl,
				"During your turn, spells you cast cost {1} less to cast for each creature you control with power 4 or greater.",
				effects.YourSpell(), effects.ActivatedDuringYourTurn()),
		},
	})
}
