package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boros Fury-Shield — Instant {2}{W}:
//
//	"Prevent all combat damage that would be dealt by target attacking or blocking creature this turn. If {R} was spent to cast this spell, Boros Fury-Shield deals damage to that creature's controller equal to the creature's power."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the spell resolves (CR 400.7). The {R}
// rider reads the spell's own payment (StackItem.Paid, Ribbons of Night's
// read), and the damage is dealt by the spell, not the creature (its
// ruling), equal to the creature's power as the spell resolves.
//
// Caveat: with strict mana off the engine does not see the mana paid, so
// the rider never happens. That is the weaker direction.
func init() {
	Register(Spec{
		OracleID:     "20eff2ce-f26d-48a2-8a1f-435a7e2968c8",
		Name:         "Boros Fury-Shield",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so Boros Fury-Shield never deals the damage."},
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return shieldAgainstTheTargetThen(ctx, true, func(ctx *Context, target uuid.UUID) error {
				if ctx.ManaSpent().Count("R") == 0 {
					return nil
				}
				ctx.Game.RecomputeLayersIfStaleLocked()
				c, ok := ctx.Game.LookupCardForEffect(target)
				if !ok || c.CurrentPower() <= 0 || ctx.Game.PlayerByIDForEffect(c.Controller) == nil {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: c.Controller, Amount: c.CurrentPower()}.Apply(ctx)
			})
		},
	})
}
