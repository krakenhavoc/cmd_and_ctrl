package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanctify — Sorcery {1}{W} (Innistrad: Crimson Vow):
//
//	"Destroy target artifact or enchantment. You gain 3 life."
//
// Disenchant's target and destruction, then the life in printed order
// (CR 608.2c). A target that is no longer legal makes the spell do
// nothing, the life included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c804ac8e-7427-43c6-b496-3ee6de192f03",
		Name:         "Sanctify",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve:    destroyTargetThenGainLife(3),
	})
}
