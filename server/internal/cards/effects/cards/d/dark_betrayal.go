package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Betrayal — Instant {B} (Theros, #5561):
//
//	"Destroy target black creature."
//
// The color is read off the creature as the spell is cast and again as
// it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3ebe9ae-108f-4d23-a236-35d632889cf8",
		Name:         "Dark Betrayal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target black creature", effects.OfColor("B")),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
