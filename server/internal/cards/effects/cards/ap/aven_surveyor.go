package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "57831c07-6915-4579-b3e1-de93a3664444",
		Name:            "Aven Surveyor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{avenSurveyorEnters()},
	})
}

// avenSurveyorEnters is the modal entry trigger: a +1/+1 counter on the
// Surveyor, or return target creature to its owner's hand.
func avenSurveyorEnters() game.TriggeredAbility {
	enters := effects.WhenThisEnters("Aven Surveyor — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on this creature.", nil, effects.PutAPlusOneCounterOnThisCreatureMode),
		effects.ModeDoing("Return target creature to its owner's hand.", effects.TargetCreature("target creature"), effects.BounceTheModesTarget),
	)
	return enters
}
