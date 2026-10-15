package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carrion Wurm — Creature — Zombie Wurm {3}{B}{B}, 6/5:
//
//	"Whenever this creature attacks or blocks, any player may exile three
//	 cards from their graveyard. If a player does, this creature assigns
//	 no combat damage this turn."
//
// Carrion Rats' ability with a price of three cards (ADR 0187, #3571): a
// payer with fewer than three cards in their graveyard can't pay
// (CR 118.3), and the three come from one player's own graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed0cf504-c365-4486-92bc-c329b83b99d0",
		Name:         "Carrion Wurm",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{carrionTrigger("Carrion Wurm", 3)},
	})
}
