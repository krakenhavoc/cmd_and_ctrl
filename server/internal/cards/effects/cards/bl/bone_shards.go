package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bone Shards — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 discard a card. Destroy target creature or planeswalker."
//
// The either/or cost (ADR 0100 §2), paid at CR 601.2h with the spell on
// the stack. The sacrificed creature cannot be the target: it is gone
// before the spell resolves, and CR 608.2b then finds the target
// illegal — which is the printed card's behaviour too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e760cfe-45b9-4a3f-b2e2-4ca6ec2bd13a",
		Name:         "Bone Shards",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("a creature", effects.Creature()).Keyed("sacrifice"),
			effects.DiscardCost(1).Keyed("discard"),
		),
		Targets:   effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
