package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thalakos Mistfolk — {2}{U} Creature — Thalakos Illusion, 2/1 (Tempest, S82 OLD
// batch 770, #5628):
//
//	"Shadow
//	 {U}: Put this creature on top of its owner's library."
//
// Shadow is a printed keyword. The ability is Wayward Soul's: a Mistfolk that has
// left the battlefield, or left and come back as a new object (CR 400.7), is not
// moved, and it goes on top of its OWNER's library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c836f8e4-6154-4c7a-a8dd-165ff74f9115",
		Name:            "Thalakos Mistfolk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Put this creature on top of its owner's library.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.PutThisPermanentOnTopOfOwnersLibrary,
		}},
	})
}
