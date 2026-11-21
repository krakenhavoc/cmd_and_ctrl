package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enlisted Wurm — Creature — Wurm {4}{G}{W}, 5/5 (Foundations Jumpstart):
//
//	"Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast it
//	 without paying its mana cost. Put the exiled cards on the bottom in a
//	 random order.)"
//
// The engine's cascade keyword whole, read off the spell's own mana value
// (six).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69191459-36ca-4e71-b4bb-0270c4cdaee9",
		Name:         "Enlisted Wurm",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
	})
}
