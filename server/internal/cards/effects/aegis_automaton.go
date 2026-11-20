package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aegis Automaton — Artifact Creature — Construct {2}, 0/3 (Aether Revolt,
// #5536):
//
//	"{4}{W}: Return another target creature you control to its owner's hand."
//
// "Another" is object identity (Another), so the Automaton can't pick
// itself. A target that left in response is skipped (bounceTheTargetToHand).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "727f4734-53cc-4279-ba51-dfa26c4ceb1d",
		Name:         "Aegis Automaton",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Return another target creature you control to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{4}{W}"),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect:  bounceTheTargetToHand,
		}},
	})
}
