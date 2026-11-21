package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Axe — Instant {R}:
//
//	"As an additional cost to cast this spell, discard a card or pay
//	 {5}. Lightning Axe deals 5 damage to target creature."
//
// The either/or cost (ADR 0100 §2). The "pay {5}" branch joins the total
// at CR 601.2f through the one pricer, so the auto-tapper, the preview
// and the bot all charge {5}{R} for it, and a cost modifier sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81b90905-fbc0-426a-a084-c3300533abb4",
		Name:         "Lightning Axe",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.DiscardCost(1).Keyed("discard"),
			effects.ManaAdditionalCost("{5}").Keyed("mana"),
		),
		Targets:   effects.TargetCreature("target creature"),
		Purpose:   effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve: effects.DamageToFirstTargetStep(5),
	})
}
