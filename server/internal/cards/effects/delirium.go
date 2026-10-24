package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Delirium — Instant {1}{B}{R}:
//
//	"Cast this spell only during an opponent's turn.
//	 Tap target creature that player controls. That creature deals damage
//	 equal to its power to the player. Prevent all combat damage that
//	 would be dealt to and dealt by the creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): "that player" is the opponent
// whose turn it is, so the target is a creature the active player
// controls, checked as it is chosen and again as the spell resolves
// (CR 608.2b). The creature is tapped, deals damage equal to its power
// as the spell resolves to that player, and then one to-and-by record
// (Mod.AndDealtBy) shields it from combat damage both ways for the rest
// of the turn.
//
// No simplifications.
// deliriumTarget is the creature Delirium taps, read from the item each step.
func deliriumTarget(ctx *Context) (uuid.UUID, bool) {
	ts := ctx.LegalTargets()
	if len(ts) == 0 || ts[0].Kind != game.TargetCard {
		return uuid.Nil, false
	}
	return ts[0].ID, true
}

func init() {
	Register(Spec{
		OracleID:     "a00e5eb6-b916-4cd6-8ff2-c02f4929abc7",
		Name:         "Delirium",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			active := activePlayerIDOf(g)
			return active != uuid.Nil && active != controller
		},
		CastConditionLabel: "Cast this spell only during an opponent's turn.",
		Targets: TargetCreature("target creature that player controls", func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			active := activePlayerIDOf(g)
			return active != uuid.Nil && c.Controller == active
		}),
		// Each sentence waits for the prompt the one before asks (a CR 616 order
		// on the damage, a replacement offer on the tap; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := deliriumTarget(ctx)
				if !ok {
					return nil
				}
				return TapTarget{Target: id}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				id, ok := deliriumTarget(ctx)
				if !ok {
					return nil
				}
				ctx.Game.RecomputeLayersIfStaleLocked()
				if c, ok := ctx.Game.LookupCardForEffect(id); ok {
					if active := activePlayerIDOf(ctx.Game); active != uuid.Nil {
						return DealDamage{Source: id, Target: active, Amount: c.CurrentPower()}.Apply(ctx)
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				id, ok := deliriumTarget(ctx)
				if !ok {
					return nil
				}
				return toAndByShield(ShieldObject(id), true).Apply(ctx)
			}),
		),
	})
}
