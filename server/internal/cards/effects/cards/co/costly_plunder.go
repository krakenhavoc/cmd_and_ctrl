package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Costly Plunder — Instant {1}{B} (Ixalan):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Draw two cards."
//
// Rite of Oblivion's additional cost over an artifact or a creature: paid as
// the spell goes on the stack (CR 118.8), so a caster with neither can't cast
// it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "8cdfd9df-4783-4138-a286-828a97f516aa",
		Name:           "Costly Plunder",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:        game.Purpose{Draws: 2},
		OnResolve:      effects.ResolveSteps(effects.DrawCards{N: 2}),
	})
}
