package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overload (Invasion, S82 OLD batch 648, #5616):
//
//	Instant {R}:
//	"Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	Destroy target artifact if its mana value is 2 or less. If this spell was kicked, destroy
//	that artifact if its mana value is 5 or less instead."
func init() {
	effects.Register(effects.Spec{
		OracleID:      "07159efc-c69f-4164-a8ca-9da641dbf702",
		Name:          "Overload",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}")},
		Targets:       effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve:     effects.S82OldB648OverloadResolve,
	})
}
