package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cursebreak — {1}{W} Instant (Avacyn Restored, #5570):
//
//	"Destroy target enchantment. You gain 2 life."
//
// The shared destroyTargetThenGainLife body (Sanctify) for an
// enchantment and 2 life: an illegal target means no life either
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac5a13b1-0725-4978-8c93-cebd76be58ba",
		Name:         "Cursebreak",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve:    destroyTargetThenGainLife(2),
	})
}
