package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Natural End — {2}{G} Instant (Avacyn Restored, #5570):
//
//	"Destroy target artifact or enchantment. You gain 3 life."
//
// Sanctify's text, the shared destroyTargetThenGainLife body: an
// illegal target means no life either (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05b5c0fb-84b2-4a32-bf66-179c4dccb3df",
		Name:         "Natural End",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve:    destroyTargetThenGainLife(3),
	})
}
