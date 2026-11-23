package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raking Claws — Instant {1}{R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gains double strike until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Assault Strobe's grant, ending at cleanup, and cycling from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2248d3d-3157-4610-868d-2b3fae2b1ec8",
		Name:         "Raking Claws",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")), Answers: game.AnswerCombatGrant},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.IkoTargetGainsUntilEOT("double strike", "Raking Claws — double strike"),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
