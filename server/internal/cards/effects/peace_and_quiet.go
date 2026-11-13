package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peace and Quiet — Instant {1}{W} (Urza's Legacy, #5623):
//
//	"Destroy two target enchantments."
//
// Two different enchantments (CR 115.3 forbids one object twice for one
// "target" word), destroyed together as one event by destroyEachLegalTarget.
// One that left or stopped being an enchantment in response is skipped, and
// the other is still destroyed (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0517f736-e6eb-4fdb-8288-e0c8a9a44379",
		Name:         "Peace and Quiet",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("two target enchantments", Enchantment()).WithCount(2, 2),
		OnResolve:    destroyEachLegalTarget,
	})
}
