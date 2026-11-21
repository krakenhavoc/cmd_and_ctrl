package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Game Over — {3}{B}{B} Sorcery (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"This spell costs {2} less to cast if a player's life total is less
//	 than or equal to half their starting life total.
//	 Destroy all creatures."
//
// Wrath of God under a self cost reduction (CR 601.2f): the reduction
// touches only the generic part, {3}{B}{B} becoming {1}{B}{B}, and is
// judged as the spell is cast. Any player still in the game counts, the
// caster included; the starting total is the table's (40 in Commander).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e326b5a1-f8ce-471b-8663-ec1ba649b24f",
		Name:         "Game Over",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if a player's life total is less than or equal to half their starting life total.", effects.APlayerIsAtOrBelowHalfTheirStartingLife()),
		},
		OnResolve: effects.WrathDestroyAllCreatures,
	})
}
