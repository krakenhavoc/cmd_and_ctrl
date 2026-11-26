package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burn the Impure — Instant {1}{R} (Mirrodin Besieged, S82 OLD b467, #5576):
//
//	"Burn the Impure deals 3 damage to target creature. If that creature has
//	 infect, Burn the Impure deals 3 damage to that creature's controller."
//
// Two instructions in printed order. Whether the creature has infect, and who
// controls it, are read as it was just before the damage (CR 608.2h), so a
// creature the first 3 damage kills still counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9503c801-eb6f-4f34-9062-c56105509b7b",
		Name:         "Burn the Impure",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			ctx.Game.RecomputeLayersIfStaleLocked()
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			infect, controller := game.HasKeyword(&c, "infect"), c.Controller
			if err := effects.DamageToFirstTargetStep(3)(item, ctx); err != nil {
				return err
			}
			if !infect || ctx.Game.PlayerByIDForEffect(controller) == nil {
				return nil
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: controller, Amount: 3}.Apply(ctx)
		},
	})
}
