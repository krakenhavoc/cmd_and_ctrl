package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Give // Take — split card, oracle 67a82d5e-a9dc-4e8c-9d23-c6d6aa1e57d1
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Give — Sorcery {2}{G}: "Put three +1/+1 counters on target creature."
//	Take — Sorcery {2}{U}: "Remove all +1/+1 counters from target creature you
//	       control. Draw that many cards."
//	       Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Give under the bare oracle ID, Take under
// "#1"); the fused cast is the engine's (CR 702.102), Give resolving first, so a
// fused Take on the same creature draws the three Give added. Take counts the
// +1/+1 counters on the creature as it resolves and draws exactly the number
// removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67a82d5e-a9dc-4e8c-9d23-c6d6aa1e57d1",
		Name:         "Give",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.S82OldB402CountersOnTheTarget(game.CounterPlusOne, 3),
	})
	effects.Register(effects.Spec{
		OracleID:     "67a82d5e-a9dc-4e8c-9d23-c6d6aa1e57d1#1",
		Name:         "Take",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			n := c.Counters[game.CounterPlusOne]
			if n <= 0 {
				return nil
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: -n}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: n}.Apply(ctx)
		},
	})
}
