package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mob — Instant {4}{B} (Modern Horizons, #5519):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you tap
//	 while casting this spell pays for {1} or one mana of that creature's
//	 color.)
//	 Destroy target creature."
//
// Convoke is the shared tap cost; the destroy is destroyTheTargetPermanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5307a61f-2142-4e12-8329-ceb729f6eb79",
		Name:         "Mob",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		TapCost:      effects.Convoke(),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
