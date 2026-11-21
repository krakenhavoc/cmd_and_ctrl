package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spark Harvest — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 pay {3}{B}. Destroy target creature or planeswalker."
//
// The either/or cost (ADR 0100 §2); the {3}{B} branch joins the total at
// CR 601.2f through the one pricer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25756e33-25e5-4976-8cce-a39fa2c4b007",
		Name:         "Spark Harvest",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("a creature", effects.Creature()).Keyed("sacrifice"),
			effects.ManaAdditionalCost("{3}{B}").Keyed("mana"),
		),
		Targets:   effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
