package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prohibit — Instant {1}{U} (Modern Horizons):
//
//	"Kicker {2}
//	 Counter target spell if its mana value is 2 or less. If this spell was
//	 kicked, counter that spell if its mana value is 4 or less instead."
//
// The mana value is read as Prohibit resolves, off the spell it targets (CR
// 608.2b: the target is still legal whatever its mana value, the rider only
// decides whether it is countered), and the limit is 4 for a kicked Prohibit
// and 2 for an unkicked one. A spell whose cost the engine can't read is never
// countered rather than counted as zero.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d95f1797-56f9-41be-a5fe-8398961f4b8b",
		Name:          "Prohibit",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}")},
		Targets:       effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			limit := 2
			if ctx.WasKicked() {
				limit = 4
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				spell, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				if mv, ok := ctx.Game.ManaValueForEffect(spell); !ok || mv > limit {
					return nil
				}
				return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
