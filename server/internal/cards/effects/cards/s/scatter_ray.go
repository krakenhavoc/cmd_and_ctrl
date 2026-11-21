package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scatter Ray — Instant for {1}{U} (The Brothers' War):
//
//	"Counter target artifact or creature spell unless its controller pays {4}."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b515dfc6-1a1e-46d6-a7a3-0e3af8279f49",
		Name:         "Scatter Ray",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target artifact or creature spell", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{4}",
				Question: "Scatter Ray — pay {4}, or your spell is countered?",
			}.Apply(ctx)
		},
	})
}
