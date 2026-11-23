package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Machine Over Matter — Instant for {1}{U} (The Brothers' War):
//
//	"This spell costs {1} less to cast if you control an artifact creature.
//	Return target nonland permanent to its owner's hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37069c66-c17a-40ac-80f4-14ef49d18008",
		Name:         "Machine Over Matter",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control an artifact creature.",
				effects.YouControlA(effects.And(effects.Artifact(), effects.Creature()))),
		},
		OnResolve: effects.BounceTheTarget,
	})
}
