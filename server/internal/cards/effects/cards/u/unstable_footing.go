package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unstable Footing — Instant {R}:
//
//	"Kicker {3}{R}
//	 Damage can't be prevented this turn. If this spell was kicked, it
//	 deals 5 damage to target player or planeswalker."
//
// The target exists only when the spell is kicked (WhenPaid on the
// kicker, #1716): an unkicked Unstable Footing targets nothing and is
// just the turn grant (ADR 0107 §5). A kicked one whose target becomes
// illegal is countered on resolution (CR 608.2b), grant and all.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb901095-2ea4-4b9e-ad5a-2892894dc4c3",
		Name:         "Unstable Footing",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			effects.WhenPaid(effects.Kicker("{3}{R}"), effects.TargetPlayerOrPlaneswalker()),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DamageCantBePreventedThisTurn{}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return effects.DamageToFirstTargetStep(5)(item, ctx)
		},
	})
}
