package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Banishing — Instant {2}{B} (Eighth Edition, #5607):
//
//	"Destroy target nonblack creature. It can't be regenerated."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0df450be-9bd2-45be-ae59-fa8d19f6a391",
		Name:         "Dark Banishing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}
