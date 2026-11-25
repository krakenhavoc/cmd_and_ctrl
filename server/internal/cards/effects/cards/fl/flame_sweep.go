package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Sweep — Instant {2}{R} (Core Set 2020):
//
//	"Flame Sweep deals 2 damage to each creature except for creatures you
//	 control with flying."
//
// Anger of the Gods' sweep with a spared class: a creature you control that
// has flying through the layers is left out, everything else is dealt 2
// damage at once (CR 120.3). The sweep is a partial one for the bots.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b34f471-40da-4c7a-81df-245fc3d14092",
		Name:         "Flame Sweep",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.Not(effects.And(effects.YouControl(), effects.HasKeyword("flying")))), 2)
		},
	})
}
