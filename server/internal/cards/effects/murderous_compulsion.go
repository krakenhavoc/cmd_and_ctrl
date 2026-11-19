package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murderous Compulsion — Sorcery {1}{B} (Innistrad Remastered):
//
//	"Destroy target tapped creature.
//	 Madness {1}{B}"
func init() {
	Register(Spec{
		OracleID:     "8fd8b4ef-69c8-43aa-9cbe-c46dba845ba8",
		Name:         "Murderous Compulsion",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Madness:      "{1}{B}",
		Targets:      TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
