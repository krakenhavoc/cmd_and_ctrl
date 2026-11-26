package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incremental Growth — Khans of Tarkir (S82 OLD b346, #5554):
//
//	Sorcery {3}{G}{G}: Put a +1/+1 counter on target creature, two +1/+1 counters on
//	another target creature, and three +1/+1 counters on a third target creature.
//
// The three clauses are Distinct, so no creature can be picked twice; a clause
// whose creature is gone by resolution is skipped and the others still resolve
// (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "802ff367-e52e-4f79-aa0f-96c45fe477a0",
		Name:         "Incremental Growth",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("target creature").Then(
			effects.Distinct(effects.TargetCreature("another target creature")),
			effects.Distinct(effects.TargetCreature("a third target creature"))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for i, t := range item.Targets {
				if i > 2 || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: i + 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
