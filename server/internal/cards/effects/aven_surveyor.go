package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aven Surveyor — Creature — Bird Scout {3}{U}{U}, 2/2 (Fate Reforged,
// #5552):
//
//	"Flying
//	 When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature.
//	 • Return target creature to its owner's hand."
//
// Flying is printed. The modal entry trigger is Knight of Autumn's shape (CR
// 700.2): the bounce bullet is offered only when there is a creature to
// target, and it may target the Surveyor itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "57831c07-6915-4579-b3e1-de93a3664444",
		Name:            "Aven Surveyor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{avenSurveyorEnters()},
	})
}

// avenSurveyorEnters is the modal entry trigger: a +1/+1 counter on the
// Surveyor, or return target creature to its owner's hand.
func avenSurveyorEnters() game.TriggeredAbility {
	enters := WhenThisEnters("Aven Surveyor — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on this creature.", nil, putAPlusOneCounterOnThisCreatureMode),
		ModeDoing("Return target creature to its owner's hand.", TargetCreature("target creature"), BounceTheModesTarget),
	)
	return enters
}
