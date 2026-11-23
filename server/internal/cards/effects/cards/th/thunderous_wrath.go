package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderous Wrath — Instant {4}{R}{R} (#1665):
//
//	"Thunderous Wrath deals 5 damage to any target.
//	 Miracle {R} (You may cast this card for its miracle cost when you
//	 draw it if it's the first card you drew this turn.)"
//
// Lightning Bolt's body at five, and six mana for it cast normally.
// Already an instant, so the miracle cast needs nothing the grant's
// timing adds — it is the price that matters.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "78260893-c443-44c8-ab45-ce86ef347d98",
		Name:             "Thunderous Wrath",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetAny(),
		Purpose:          effects.ForTargets(effects.DamageToTarget(0, 5)),
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{R}")},
		OnResolve:        effects.DamageToFirstTargetStep(5),
	})
}
