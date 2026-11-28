package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leeches — Sorcery {1}{W}{W} (Homelands, S82 OLD b824, #5635):
//
//	"Target player loses all poison counters. Leeches deals that much damage to
//	that player."
//
// The counters are removed first and the damage is the number removed (CR
// 122.1, 608.2c): a player with none loses none and takes none. The removal is
// a counter change on the player, so a replacement may alter it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a265cb5-47c7-4e06-9030-602e17bedae5",
		Name:         "Leeches",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
				return nil
			}
			p := ctx.Game.PlayerByIDForEffect(t.ID)
			if p == nil || p.Eliminated {
				return nil
			}
			n := p.Counters[game.CounterPoison]
			if n <= 0 {
				return nil
			}
			if err := ctx.Game.AddPlayerCounterByForEffect(ctx.Source(), t.ID, game.CounterPoison, -n); err != nil {
				return err
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		},
	})
}
