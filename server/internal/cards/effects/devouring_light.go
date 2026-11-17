package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devouring Light — Instant {1}{W}{W} (Ravnica: City of Guilds, #5600):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Exile target attacking or blocking creature."
//
// Convoke is Spec.TapCost (CR 702.51), so a white creature can pay a {W}.
// The target must still be attacking or blocking as the spell resolves; a
// creature removed from combat in response is not exiled (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5334dc24-040e-4094-80f4-8d1b9de01230",
		Name:         "Devouring Light",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		TapCost:      Convoke(),
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve:    ResolveSteps(Instruction(exileTheFirstTargetIfLegal)),
	})
}
