package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jhoira's Toolbox — Artifact Creature — Insect {2}, 1/1 (Urza's Legacy, S82 OLD b716, #5623):
//
//	"{2}: Regenerate target artifact creature."
//
// Boneknitter's regeneration shield for an artifact creature of any player's,
// the Toolbox included (it is one); the target is read through the layers at
// announce and again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a08475a6-1fac-4a3f-b343-aed16a61bc16",
		Name:         "Jhoira's Toolbox",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Regenerate target artifact creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
