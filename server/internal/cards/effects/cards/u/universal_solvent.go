package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Universal Solvent — Artifact {1} (Foundations Jumpstart):
//
//	"{7}, {T}, Sacrifice this artifact: Destroy target permanent."
//
// Aetherjacket's destroy ability with the cost and without the sorcery
// restriction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f34ba913-cf9a-48fe-b5fe-68bf504693b2",
		Name:         "Universal Solvent",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}, {T}, Sacrifice this artifact: Destroy target permanent.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{7}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
