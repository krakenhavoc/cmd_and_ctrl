package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fadf984c-c224-432a-87b9-b34f929b8409",
		Name:         "Temur Battlecrier",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLessEach(stdB9BigCreaturesYouControl,
				"During your turn, spells you cast cost {1} less to cast for each creature you control with power 4 or greater.",
				YourSpell(), ActivatedDuringYourTurn()),
		},
	})
}
