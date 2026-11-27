package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stall for Time — Instant {2}{W} (Dominaria United):
//
//	"Kicker {1}{U} (You may pay an additional {1}{U} as you cast this
//	 spell.)
//	 Tap up to two target creatures. If this spell was kicked, put a
//	 stun counter on each of those creatures. (If a permanent with a
//	 stun counter would become untapped, remove one from it instead.)
//	 Draw a card."
//
// Printed order is resolution order (CR 608.2c): tap, then the kicked
// stun counters on the same creatures, then the draw. "Up to two" may
// name none, and the spell still draws. A target that has left is
// skipped (CR 608.2b); if every target has left the spell does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "09a400bd-83f1-4fac-a218-64c9f041874f",
		Name:          "Stall for Time",
		Purpose:       game.Purpose{Draws: 1},
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{U}")},
		Targets:       effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.OldB711TapTheTargets),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.DrawCards{N: 1},
		),
	})
}
