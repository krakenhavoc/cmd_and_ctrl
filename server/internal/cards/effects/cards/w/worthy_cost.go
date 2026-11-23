package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worthy Cost — Sorcery {B} (Tarkir: Dragonstorm, #4463):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Exile target creature or planeswalker."
//
// Altar's Reap's mandatory sacrifice as the additional cost, and Baleful
// Mastery's exile of a creature or planeswalker. The spell can't be cast
// without a creature to sacrifice. The target is chosen before the cost is
// paid (CR 601.2c, CR 601.2h), so a creature you sacrifice is no longer a
// legal target and a spell whose only target is gone does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0f8ebf11-b5ec-4f21-8c49-90d569b5a4c5",
		Name:           "Worthy Cost",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTheFirstTargetIfLegal(ctx)
		},
	})
}
