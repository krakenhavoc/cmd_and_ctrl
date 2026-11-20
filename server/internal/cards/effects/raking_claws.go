package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raking Claws — Instant {1}{R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gains double strike until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Assault Strobe's grant, ending at cleanup, and cycling from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2248d3d-3157-4610-868d-2b3fae2b1ec8",
		Name:         "Raking Claws",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "double strike")), Answers: game.AnswerCombatGrant},
		Targets:      TargetCreature("target creature"),
		OnResolve:    ikoTargetGainsUntilEOT("double strike", "Raking Claws — double strike"),
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
