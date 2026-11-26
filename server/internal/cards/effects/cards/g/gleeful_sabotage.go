package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gleeful Sabotage — {1}{G} Sorcery (Archenemy):
//
//	"Destroy target artifact or enchantment.
//	 Conspire"
func init() {
	effects.Register(effects.Spec{
		OracleID:      "25fe48be-95c4-4011-8c60-4628fb5ecfcb",
		Name:          "Gleeful Sabotage",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		Targets:       effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve:     effects.DestroyTheTargetPermanent,
	})
}
