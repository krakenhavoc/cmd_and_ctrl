package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Johann's Stopgap — Sorcery {3}{U} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 This spell costs {2} less to cast if it's bargained.
//	 Return target nonland permanent to its owner's hand. Draw a card."
//
// Bargain is the shared optional cost (g2Bargain) and the {2} discount
// reads the announcement as the total cost is fixed
// (CostsLessIfBargained, CR 601.2f), so a bargained Stopgap costs {1}{U}
// and its mana value stays 4 (CR 202.3). The body is Fantastic Bounce's
// (bounceTheTargetThenDraw): the bounce, then the draw as a keyed tail,
// which draws whether or not the permanent reached its owner's hand (a
// commander may go to the command zone instead); a target gone in
// response counters the spell, draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "c7b235e1-746e-4467-be54-4e11e2bf6362",
		Name:          "Johann's Stopgap",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Draws: 1},
		Targets:       TargetPermanent("target nonland permanent", Nonland()),
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfBargained(2, "This spell costs {2} less to cast if it's bargained."),
		},
		OnResolve: bounceTheTargetThenDraw,
	})
}
