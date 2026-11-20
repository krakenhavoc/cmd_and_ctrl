package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solemn Offering — Sorcery {2}{W} (Magic 2010, #5585):
//
//	"Destroy target artifact or enchantment. You gain 4 life."
//
// Sanctify's shape with 4 life: the destruction, then the life. A target
// that is no longer legal makes the spell do nothing, the life included
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aff51c27-fb61-41b3-a7ea-7092cd280c63",
		Name:         "Solemn Offering",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve:    destroyTargetThenGainLife(4),
	})
}
