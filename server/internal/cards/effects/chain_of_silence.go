package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chain of Silence — Instant {1}{W}:
//
//	"Prevent all damage target creature would deal this turn. That creature's controller may sacrifice a land of their choice. If the player does, they may copy this spell and may choose a new target for that copy."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the spell resolves (CR 400.7), all of its
// damage for the rest of the turn. The rest is the Chain cycle's sentence
// (chain_of_vapor.go): the creature's controller, read as the spell
// resolves, may sacrifice a land, and then may copy the spell under their
// own control with a new target (CR 707.10, 707.10b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d7b476ed-fc50-4804-8375-488a9e2ab184",
		Name:         "Chain of Silence",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return shieldAgainstTheTargetThen(ctx, false, func(ctx *Context, target uuid.UUID) error {
				c, ok := ctx.Game.LookupCardForEffect(target)
				if !ok {
					return nil
				}
				return chainSacrificeALandToCopy(ctx, c.Controller, "Chain of Silence")
			})
		},
	})
}
