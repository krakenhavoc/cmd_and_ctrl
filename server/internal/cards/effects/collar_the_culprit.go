package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collar the Culprit — Instant {3}{W} (Guilds of Ravnica, #5525):
//
//	"Destroy target creature with toughness 4 or greater."
//
// Sungold Barrage's body. Toughness is read as the spell is cast and again
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "adc03c37-33c7-4b92-b9fd-4608e6c372c4",
		Name:         "Collar the Culprit",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with toughness 4 or greater", ToughnessGE(4)),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    s82MidDestroyTheTarget,
	})
}
