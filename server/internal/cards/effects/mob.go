package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5307a61f-2142-4e12-8329-ceb729f6eb79",
		Name:         "Mob",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		TapCost:      Convoke(),
		Targets:      TargetCreature("target creature"),
		OnResolve:    destroyTheTargetPermanent,
	})
}
