package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rowdy Research — Instant {6}{U} (Wilds of Eldraine, #4471):
//
//	"This spell costs {1} less to cast for each creature that attacked
//	 this turn.
//	 Draw three cards."
//
// The Mary Janes' discount (creaturesThatAttackedThisTurn): distinct
// creatures declared as attackers this turn, anyone's, through the one
// pricer (CR 601.2f). It reduces generic mana only, so the spell never
// costs less than {U}. Then Arcane Epiphany's draw of three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a33a5c2-5aa1-48e5-95a8-d52cdce980de",
		Name:         "Rowdy Research",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(creaturesThatAttackedThisTurn,
				"This spell costs {1} less to cast for each creature that attacked this turn"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 3}.Apply(ctx)
		},
	})
}
