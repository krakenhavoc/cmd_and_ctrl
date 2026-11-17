package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Circular Logic — Instant {2}{U} (Torment, #5612):
//
//	"Counter target spell unless its controller pays {1} for each card in
//	 your graveyard.
//	 Madness {U} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Logic Knot's CounterUnlessPaid with the price counted instead of
// announced: the cards in the graveyard of Circular Logic's controller,
// counted once as it resolves (CR 608.2h). Circular Logic is on the stack
// then, so it does not count itself. An empty graveyard asks for {0},
// which the spell's controller can always pay. A target that is no
// longer legal makes the spell do nothing (CR 608.2b).
//
// Madness is the engine's (CR 702.35a, #657), grown from the one string,
// as Obsessive Search's: a discarded Circular Logic is exiled and offered
// for {U}, at instant speed, with its target chosen as it is cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b071bbaf-79fa-413b-807b-2cc86c8a8125",
		Name:         "Circular Logic",
		Completeness: CompletenessFull,
		Madness:      "{U}",
		Targets:      TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			cost := "{" + strconv.Itoa(b31GraveyardSize(ctx.Game, ctx.Controller())) + "}"
			return CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     cost,
				Question: "Circular Logic — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
