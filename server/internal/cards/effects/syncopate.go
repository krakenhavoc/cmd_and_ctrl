package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Syncopate — Instant {X}{U}, reprinted in Final Fantasy (#4462):
//
//	"Counter target spell unless its controller pays {X}. If that spell
//	 is countered this way, exile it instead of putting it into its
//	 owner's graveyard."
//
// Logic Knot's "unless its controller pays {X}", the X read off the
// resolving spell (X = 0 asks for {0}, which the controller pays for
// nothing), with the countered spell exiled (CounterUnlessPaid.Dest,
// #5236). The destination rides the prompt as its keyed decline, so a
// table waiting on the payer is a restore point and still exiles on a
// decline. A spell that can't be countered is neither countered nor
// exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "318f7f70-e374-40ef-8afb-3389c10461d8",
		Name:         "Syncopate",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			cost := "{" + strconv.Itoa(ctx.X()) + "}"
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     cost,
				Question: "Syncopate — pay " + cost + " or your spell is exiled",
				Dest:     game.ZoneRef{Kind: game.ZoneExile},
			}.Apply(ctx)
		},
	})
}
