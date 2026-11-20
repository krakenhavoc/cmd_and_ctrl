package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murderous Compulsion — Sorcery {1}{B} (Shadows over Innistrad, #5543):
//
//	"Destroy target tapped creature.
//	 Madness {1}{B} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Death Stroke's "destroy target tapped creature"
// (destroyTheTargetPermanent). Madness is grown by the engine from the one
// string (CR 702.35a, #657).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8fd8b4ef-69c8-43aa-9cbe-c46dba845ba8",
		Name:         "Murderous Compulsion",
		Completeness: CompletenessFull,
		Madness:      "{1}{B}",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
