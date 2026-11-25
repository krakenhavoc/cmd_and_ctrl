package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consign to Dust — Instant {2}{G}:
//
//	"Strive — This spell costs {2}{G} more to cast for each target beyond the
//	 first.
//	 Destroy any number of target artifacts and/or enchantments."
//
// Strive is Rouse the Mob's self cost modifier. The destruction is one
// simultaneous event over every target still legal at resolution
// (DestroyEachLegalTarget, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ac493f2-9519-44d7-b6ac-8c020ad2cede",
		Name:         "Consign to Dust",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("any number of target artifacts and/or enchantments", effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{2}{G}", "Strive — This spell costs {2}{G} more to cast for each target beyond the first."),
		},
		OnResolve: effects.DestroyEachLegalTarget,
	})
}
