package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Poplar Shaman — Creature — Treefolk Shaman {2}{B}, 1/3 (Lorwyn,
// S82 OLD b547, #5593):
//
//	"{2}{B}: Regenerate target Treefolk."
//
// Boneknitter's regeneration for another creature type: the shield is put on
// the target as the ability resolves (CR 701.19a), and a changeling is a
// Treefolk (CR 702.73a). The Shaman is a Treefolk and may target itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39d60cd8-bebc-407d-b361-8e49e70e339a",
		Name:         "Black Poplar Shaman",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: Regenerate target Treefolk.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{2}{B}"),
			Targets: effects.TargetPermanent("target Treefolk", effects.HasSubtype("Treefolk")),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
