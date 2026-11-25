package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sonic Assault — Instant {1}{U}{R} (Guilds of Ravnica, #5525):
//
//	"Tap target creature. Sonic Assault deals 2 damage to that creature's
//	 controller.
//	 Jump-start"
//
// The creature is tapped and its controller, read as the spell resolves, takes
// 2 damage from Sonic Assault itself. Nothing happens if the target is no
// longer legal (CR 608.2b); a creature that is already tapped is still the
// target and its controller still takes the damage. Jump-start is the shared
// alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "dfe9e485-c070-4efd-8c35-69d4eb3f58e4",
		Name:             "Sonic Assault",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Cmdr2S39JumpStart("{1}{U}{R}")},
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := c.Controller
			if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if ctx.Game.PlayerByIDForEffect(controller) == nil {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: 2}.Apply(ctx)
		})),
	})
}
