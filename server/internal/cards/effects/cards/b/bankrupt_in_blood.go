package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bankrupt in Blood — Sorcery {1}{B} (Ravnica Allegiance):
//
//	"As an additional cost to cast this spell, sacrifice two creatures.
//	 Draw three cards."
//
// The two creatures are an additional cost (CR 118.8), paid as one simultaneous
// sacrifice while the spell goes on the stack, so a caster with fewer than two
// creatures can't cast it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "10a66403-606d-4814-9583-50cdcd6976e0",
		Name:           "Bankrupt in Blood",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeNCost(2, "two creatures", effects.Creature()),
		Purpose:        game.Purpose{Draws: 3},
		OnResolve:      effects.ResolveSteps(effects.DrawCards{N: 3}),
	})
}
