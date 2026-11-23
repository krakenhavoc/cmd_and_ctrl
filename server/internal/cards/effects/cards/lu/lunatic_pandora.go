package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lunatic Pandora — Legendary Artifact {1} (Final Fantasy, #4462):
//
//	"{2}, {T}: Surveil 1. (Look at the top card of your library. You may
//	 put it into your graveyard.)
//	 {6}, {T}, Sacrifice Lunatic Pandora: Destroy target nonland
//	 permanent."
//
// Forum of Amity's surveil ability and Giant's Boulder's sacrifice-to-
// destroy, which names its target on the stack: sacrificing Pandora is
// part of the cost, so the destroy still happens when it has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34c65c7f-33ab-4f91-84ce-d0c8f4fc5832",
		Name:         "Lunatic Pandora",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Surveil 1.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Effect:  effects.Do(effects.Surveil{N: 1}),
			},
			{
				Label:   "{6}, {T}, Sacrifice Lunatic Pandora: Destroy target nonland permanent.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{6}"},
				Targets: effects.TargetPermanent("target nonland permanent", effects.Not(effects.Land())),
				Effect:  effects.DestroyFirstLegalCardTarget,
			},
		},
	})
}
