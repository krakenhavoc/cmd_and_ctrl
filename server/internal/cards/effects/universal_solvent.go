package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Universal Solvent — Artifact {1} (Foundations Jumpstart):
//
//	"{7}, {T}, Sacrifice this artifact: Destroy target permanent."
//
// Aetherjacket's destroy ability with the cost and without the sorcery
// restriction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f34ba913-cf9a-48fe-b5fe-68bf504693b2",
		Name:         "Universal Solvent",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{7}, {T}, Sacrifice this artifact: Destroy target permanent.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{7}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target permanent"),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
