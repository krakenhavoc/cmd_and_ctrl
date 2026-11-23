package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiring Statuary — Artifact {3} (#3656):
//
//	"Nonartifact spells you cast have improvise."
//
// The continuous improvise grant Ironheart, Clever Champion shipped on
// (#3755, SpellsYouCastHaveImprovise), narrowed to nonartifact spells.
// The Statuary taps for its own grant like any other untapped artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78b2b68d-6e27-49f3-8373-6e19faf6f679",
		Name:         "Inspiring Statuary",
		Completeness: effects.CompletenessFull,
		GrantedTapCosts: []game.GrantedTapCost{effects.SpellsYouCastHaveImprovise(
			"Nonartifact spells you cast have improvise.", effects.SpellIs(effects.Not(effects.Artifact())))},
	})
}
