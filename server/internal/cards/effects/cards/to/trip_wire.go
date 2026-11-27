package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trip Wire — Sorcery {2}{G} (Portal Three Kingdoms, S82 OLD b713, #5622):
//
//	"Destroy target creature with horsemanship."
//
// The keyword is read through the layered view, at announce and again at
// resolution (CR 608.2b), so a creature that lost horsemanship in response
// is an illegal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5751df7a-aa9a-4a27-b7ab-445ec4bcbdee",
		Name:         "Trip Wire",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with horsemanship", effects.HasKeyword("horsemanship")),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
