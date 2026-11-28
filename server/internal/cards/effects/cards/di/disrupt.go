package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disrupt — Instant {U} (Invasion, S82 #5616):
//
//	"Counter target instant or sorcery spell unless its controller pays {1}.
//	 Draw a card."
//
// Spell Pierce's pay-unless prompt (to the spell's controller, the spell stays
// on the stack until answered) on instants and sorceries, then the draw. A
// target that has left the stack makes the whole spell do nothing, the draw
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76347258-71d5-4874-adde-ff7aa932b808",
		Name:         "Disrupt",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetSpell("target instant or sorcery spell", effects.Or(effects.Instant(), effects.Sorcery())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := ctx.ClauseTarget(0); !ok || len(item.Targets) == 0 {
				return nil
			}
			if err := (effects.CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{1}",
				Question: "Disrupt — pay {1} or your spell is countered",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
