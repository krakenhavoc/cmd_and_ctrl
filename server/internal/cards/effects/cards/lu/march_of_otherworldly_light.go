package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// March of Otherworldly Light — Instant {X}{W} (Kamigawa: Neon Dynasty):
//
//	"As an additional cost to cast this spell, you may exile any number
//	 of white cards from your hand. This spell costs {2} less to cast for
//	 each card exiled this way.
//	 Exile target artifact, creature, or enchantment with mana value X or
//	 less."
//
// The cost is March of Swirling Mist's. The target is bound to the
// announced X (WithManaValueAtMostX, re-checked on resolution, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ed06e421-480c-40c8-8c3e-1e466e5327b2",
		Name:           "March of Otherworldly Light",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.ExileAnyNumberFromHandCost("white cards", "W"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessPerExiledFromHand("{2}", "This spell costs {2} less to cast for each card exiled this way"),
		},
		Targets: effects.TargetPermanent("target artifact, creature, or enchantment with mana value X or less",
			effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())).WithManaValueAtMostX(),
		Purpose:   game.Purpose{Answers: game.AnswerRemove},
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.ExileTheFirstTargetIfLegal)),
	})
}
