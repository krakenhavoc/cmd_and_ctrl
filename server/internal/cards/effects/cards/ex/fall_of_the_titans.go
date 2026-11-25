package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fall of the Titans — Instant {X}{X}{R} (Oath of the Gatewatch, #5545):
//
//	"Surge {X}{R} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Fall of the Titans deals X damage to each of up to two targets."
//
// Surge is the shared gated alternative cost (CR 702.117a); its {X} is the
// same announced X, paid once instead of twice. The damage goes to each
// still-legal target (CR 608.2b), the spell as the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "55e564b4-7c08-4109-bd67-da5f31fe002e",
		Name:             "Fall of the Titans",
		Completeness:     effects.CompletenessFull,
		XMatters:         true,
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{X}{R}")},
		Targets:          effects.TargetAny().WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.X() <= 0 {
				return nil
			}
			return effects.B10DamageEachLegalTarget(ctx, ctx.X())
		},
	})
}
