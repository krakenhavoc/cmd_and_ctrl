package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rats of Rath — Creature — Rat {1}{B}, 2/1 (Tempest, S82 OLD, #5628):
//
//	"{B}: Destroy target artifact, creature, or land you control."
//
// Self-removal on demand: the clause names only permanents the controller
// controls, and the Rats may target themselves. The target is re-checked at
// resolution, so a permanent that changed hands in response is illegal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75458175-eeb1-4a6b-93d9-c744c93b3759",
		Name:         "Rats of Rath",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Destroy target artifact, creature, or land you control.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{B}"),
			Targets: effects.TargetPermanent("target artifact, creature, or land you control", effects.And(effects.Or(effects.Artifact(), effects.Creature(), effects.Land()), effects.YouControl())),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
