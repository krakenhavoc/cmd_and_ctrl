package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "78260893-c443-44c8-ab45-ce86ef347d98",
		Name:             "Thunderous Wrath",
		Completeness:     CompletenessFull,
		Targets:          TargetAny(),
		Purpose:          ForTargets(DamageToTarget(0, 5)),
		AlternativeCosts: []game.AlternativeCost{Miracle("{R}")},
		OnResolve:        damageToFirstTarget(5),
	})
}
