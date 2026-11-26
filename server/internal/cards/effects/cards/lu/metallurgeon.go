package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metallurgeon — {1}{W} Artifact Creature — Human Artificer 1/2 (Archenemy):
//
//	"{W}, {T}: Regenerate target artifact."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09373283-a5c1-4d8e-ad25-19cd55b1e869",
		Name:         "Metallurgeon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Regenerate target artifact.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
