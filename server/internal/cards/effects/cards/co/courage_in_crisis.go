package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Courage in Crisis — Sorcery {2}{G} (Commander Masters, #5948):
//
//	"Put a +1/+1 counter on target creature, then proliferate. (Choose any
//	 number of permanents and/or players, then give each another counter of
//	 each kind already there.)"
//
// Atomize's shape: the counter on the target (only while it is still
// legal, CR 608.2b), then the player's proliferate choice, so the counter
// just placed can be proliferated. With its only target gone the spell
// does not resolve at all, so there is no proliferate either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41b929ad-0c44-4e50-84d3-13b318035543",
		Name:         "Courage in Crisis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.PlusOneCounterOnTheTarget(ctx.Game, item); err != nil {
				return err
			}
			return effects.Proliferate{}.Apply(ctx)
		},
	})
}
