package cardsri

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rune Snag — Instant {1}{U} (Coldsnap, #5597):
//
//	"Counter target spell unless its controller pays {2} plus an additional {2}
//	 for each card named Rune Snag in each graveyard."
//
// Miscalculation's pay-or-be-countered with the amount counted as this
// resolves: Rune Snag itself is still on the stack, so it is not counted, and
// the cards in every player's graveyard are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49bf3805-cc9e-4550-8a01-171282aa49c9",
		Name:         "Rune Snag",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			n := 0
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					if c.Name == "Rune Snag" {
						n++
					}
				}
			}
			cost := "{" + strconv.Itoa(2+2*n) + "}"
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     cost,
				Question: "Rune Snag — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
