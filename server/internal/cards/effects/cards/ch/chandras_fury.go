package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra's Fury — Instant {4}{R} (Magic Origins):
//
//	"Chandra's Fury deals 4 damage to target player or planeswalker and
//	 1 damage to each creature that player or that planeswalker's
//	 controller controls."
//
// Lavalanche's shape with a fixed 4 for the target and 1 for the
// creatures. The creatures are read as the spell resolves; all the
// damage is dealt at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb16f188-fe12-4a13-9947-d43cd6ce270c",
		Name:         "Chandra's Fury",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose: game.Purpose{
			Targets: game.ForTargets(effects.DamageToTarget(0, 4)),
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			t := targets[0]
			player := t.ID
			if t.Kind == game.TargetCard {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				player = c.Controller
			}
			var creatures []uuid.UUID
			for _, c := range effects.MatchingBattlefield(ctx, effects.Creature()) {
				if c.Controller == player {
					creatures = append(creatures, c.InstanceID)
				}
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}).Apply(ctx); err != nil {
					return err
				}
				for _, id := range creatures {
					if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
