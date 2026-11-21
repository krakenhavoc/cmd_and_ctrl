package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inquisitor's Snare — Instant {1}{W}:
//
//	"Prevent all damage target attacking or blocking creature would deal this turn. If that creature is black or red, destroy it."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the spell resolves (CR 400.7). The pin
// keeps applying after the creature leaves the battlefield, to damage it
// deals as it last existed there (its ruling: "The prevention effect
// applies even after the creature leaves the battlefield"). Its colour
// is read as the spell resolves, after the shield is made.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55e3dfcb-0999-401a-ae59-9804e12f4cfe",
		Name:         "Inquisitor's Snare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ShieldAgainstTheTargetThen(ctx, false, func(ctx *effects.Context, target uuid.UUID) error {
				c, ok := ctx.Game.LookupCardForEffect(target)
				if !ok || !(c.HasColor(ctx.TextColor("B")) || c.HasColor(ctx.TextColor("R"))) {
					return nil
				}
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			})
		},
	})
}
