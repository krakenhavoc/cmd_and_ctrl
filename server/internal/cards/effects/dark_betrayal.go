package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Betrayal — Instant {B} (Theros, #5561):
//
//	"Destroy target black creature."
//
// The color is read off the creature as the spell is cast and again as
// it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3ebe9ae-108f-4d23-a236-35d632889cf8",
		Name:         "Dark Betrayal",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target black creature", OfColor("B")),
		OnResolve:    destroyTheTargetPermanent,
	})
}
