package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "0f8ebf11-b5ec-4f21-8c49-90d569b5a4c5",
		Name:           "Worthy Cost",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTheFirstTargetIfLegal(ctx)
		},
	})
}
