package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thraben Gargoyle // Stonewing Antagonizer — Artifact Creature — Gargoyle
// {1}, 2/2 // Artifact Creature — Gargoyle Horror 4/2 (Shadows over
// Innistrad, S82 OLD):
//
//	Front: "Defender
//	        {6}: Transform this creature."
//	Back:  "Flying"
//
// Concealing Curtains' transform activation without the sorcery gate.
//
// No simplification.
func init() {
	const oracle = "9b5cf592-8385-4115-b27a-98f42bdf6329"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Thraben Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{6}"),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Stonewing Antagonizer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
