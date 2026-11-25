package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Radiating Lightning — Instant {3}{R} (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"Radiating Lightning deals 3 damage to target player and 1 damage to each
//	 creature that player controls."
//
// The 3 to the player, then Balefire Dragon's each-creature-that-player-controls
// sweep at 1, the set read as the sweep begins.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9794f397-dd86-4cfe-a431-5f8d8c6b68ea",
		Name:         "Radiating Lightning",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Targets: game.ForTargets(effects.DamageToTarget(0, 3)),
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true},
		},
		Targets: effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Target: victim, Amount: 3, Source: item.SourceCardID}).Apply(ctx); err != nil {
				return err
			}
			return effects.B11DamageEachCreatureControlledBy(ctx.Game, item, victim, 1)
		},
	})
}
