package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
func deliriumTarget(ctx *effects.Context) (uuid.UUID, bool) {
	ts := ctx.LegalTargets()
	if len(ts) == 0 || ts[0].Kind != game.TargetCard {
		return uuid.Nil, false
	}
	return ts[0].ID, true
}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "a00e5eb6-b916-4cd6-8ff2-c02f4929abc7",
		Name:         "Delirium",
		Completeness: effects.CompletenessFull,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			active := effects.ActivePlayerIDOf(g)
			return active != uuid.Nil && active != controller
		},
		CastConditionLabel: "Cast this spell only during an opponent's turn.",
		Targets: effects.TargetCreature("target creature that player controls", func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			active := effects.ActivePlayerIDOf(g)
			return active != uuid.Nil && c.Controller == active
		}),
		// Each sentence waits for the prompt the one before asks (a CR 616 order
		// on the damage, a replacement offer on the tap; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := deliriumTarget(ctx)
				if !ok {
					return nil
				}
				return effects.TapTarget{Target: id}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := deliriumTarget(ctx)
				if !ok {
					return nil
				}
				ctx.Game.RecomputeLayersIfStaleLocked()
				if c, ok := ctx.Game.LookupCardForEffect(id); ok {
					if active := effects.ActivePlayerIDOf(ctx.Game); active != uuid.Nil {
						return effects.DealDamage{Source: id, Target: active, Amount: c.CurrentPower()}.Apply(ctx)
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := deliriumTarget(ctx)
				if !ok {
					return nil
				}
				return effects.ToAndByShield(effects.ShieldObject(id), true).Apply(ctx)
			}),
		),
	})
}
