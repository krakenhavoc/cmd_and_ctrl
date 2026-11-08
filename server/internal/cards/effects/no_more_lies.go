package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// No More Lies — Instant {W}{U}, Murders at Karlov Manor (#4469):
//
//	"Counter target spell unless its controller pays {3}. If that spell
//	 is countered this way, exile it instead of putting it into its
//	 owner's graveyard."
//
// Spell Pierce's pay-unless at {3} on any spell, with the countered
// spell exiled (CounterUnlessPaid.Dest, #5236). The destination rides
// the prompt as its keyed decline, so a table waiting on the payer is a
// restore point and still exiles on a decline. A spell that can't be
// countered is neither countered nor exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ead1ee6a-e0da-46d5-89e8-31e8c0270bc2",
		Name:         "No More Lies",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{3}",
				Question: "No More Lies — pay {3} or your spell is exiled",
				Dest:     game.ZoneRef{Kind: game.ZoneExile},
			}.Apply(ctx)
		},
	})
}
