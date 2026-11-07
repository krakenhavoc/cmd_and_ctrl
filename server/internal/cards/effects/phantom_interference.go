package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0db20037-35b3-4b1f-8dd6-3ee797c421ca",
		Name:         "Phantom Interference",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Create a 2/2 white Spirit creature token with flying.", "{3}", nil,
				func(item *game.StackItem, ctx *Context, occ int) error {
					return CreateToken{Template: TokenCard("2/2 white Spirit with flying"), N: 1}.Apply(ctx)
				}),
			SpreeModeDoing("Counter target spell unless its controller pays {2}.", "{1}",
				TargetSpell("target spell"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || ctx.Game.StackItemForEffect(t.ID) == nil {
						return nil
					}
					return CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{2}",
						Question: "Phantom Interference — pay {2} or your spell is countered",
					}.Apply(ctx)
				}),
		),
	})
}
