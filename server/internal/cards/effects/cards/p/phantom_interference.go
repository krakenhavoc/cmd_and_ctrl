package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantom Interference — Instant {U} (Outlaws of Thunder Junction,
// #4468):
//
//	"Spree (Choose one or more additional costs.)
//	 + {3} — Create a 2/2 white Spirit creature token with flying.
//	 + {1} — Counter target spell unless its controller pays {2}."
//
// Three Steps Ahead's Spree (CR 702.172a): each bullet is its own
// closure with its own cost, and only the counter bullet targets. The
// counter is Dazzling Denial's CounterUnlessPaid, asked of the spell's
// controller on resolution. With both bullets chosen the token is made
// first, in printed order; a spell that has left the stack by then
// leaves the counter bullet nothing to do, and the token is still made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0db20037-35b3-4b1f-8dd6-3ee797c421ca",
		Name:         "Phantom Interference",
		Completeness: effects.CompletenessFull,
		Modes: effects.Spree(
			effects.SpreeModeDoing("Create a 2/2 white Spirit creature token with flying.", "{3}", nil,
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					return effects.CreateToken{Template: effects.TokenCard("2/2 white Spirit with flying"), N: 1}.Apply(ctx)
				}),
			effects.SpreeModeDoing("Counter target spell unless its controller pays {2}.", "{1}",
				effects.TargetSpell("target spell"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || ctx.Game.StackItemForEffect(t.ID) == nil {
						return nil
					}
					return effects.CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{2}",
						Question: "Phantom Interference — pay {2} or your spell is countered",
					}.Apply(ctx)
				}),
		),
	})
}
