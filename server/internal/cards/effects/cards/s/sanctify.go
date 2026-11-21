package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c804ac8e-7427-43c6-b496-3ee6de192f03",
		Name:         "Sanctify",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.DestroyTargetThenGainLife(3),
	})
}
