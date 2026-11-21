package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bayou Groff — Creature — Plant Dog {1}{G}, 5/4:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 pay {3}."
//
// The either/or cost (ADR 0100 §2) and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e83b2790-3cec-429c-a65d-6f4c7d025d37",
		Name:         "Bayou Groff",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("a creature", effects.Creature()).Keyed("sacrifice"),
			effects.ManaAdditionalCost("{3}").Keyed("mana"),
		),
	})
}
