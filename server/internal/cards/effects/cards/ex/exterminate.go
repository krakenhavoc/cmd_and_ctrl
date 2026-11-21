package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exterminate! — Sorcery {2}{B} (Doctor Who):
//
//	"Replicate—Tap an untapped Dalek you control. (When you cast this
//	 spell, copy it for each time you paid its replicate cost. You may choose
//	 new targets for the copies.)
//	 Destroy target creature. Its controller loses 3 life."
//
// Replicate's cost is tapping an untapped Dalek (ReplicateTapping, Psionic
// Ritual's shape: tapping is not the {T} symbol, so a Dalek that just arrived
// pays). The body is Bitter Downfall's "destroy, its controller loses life",
// with the controller read before the destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "90b540ab-b05a-4d29-b896-7ff5da0b101d",
		Name:          "Exterminate!",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.ReplicateTapping("an untapped Dalek you control", 10, effects.OfCreatureType("Dalek"))},
		Triggered:     []game.TriggeredAbility{effects.Replicate()},
		OnResolve:     effects.GapB95DestroyItsControllerLosesLife(3),
	})
}
