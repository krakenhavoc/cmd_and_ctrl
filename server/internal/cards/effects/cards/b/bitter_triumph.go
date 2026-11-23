package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bitter Triumph — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, discard a card or pay 3
//	 life. Destroy target creature or planeswalker."
//
// The either/or cost (ADR 0100 §2). The "pay 3 life" branch is refused
// at announce for a caster below 3 life (CR 119.4), and the view shows
// it disabled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "776341cb-d2ec-423f-9250-92dc8bd8d503",
		Name:         "Bitter Triumph",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.DiscardCost(1).Keyed("discard"),
			effects.PayLifeCost(3).Keyed("life"),
		),
		Targets:   effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
