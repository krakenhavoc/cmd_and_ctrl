package cardsr

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakshasa's Disdain — {2}{U} Instant (Fate Reforged):
//
//	"Counter target spell unless its controller pays {1} for each card in your
//	 graveyard."
//
// Circular Logic's CounterUnlessPaid with the price counted, not announced: the
// cards in the graveyard of Rakshasa's Disdain's controller, counted once as it
// resolves (CR 608.2h). The Disdain is on the stack then, so it does not count
// itself. An empty graveyard asks for {0}, which can always be paid. A target
// that is no longer legal makes the spell do nothing (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8fe6204-9f55-45e5-99b3-43f2761ce8a3",
		Name:         "Rakshasa's Disdain",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			cost := "{" + strconv.Itoa(effects.B31GraveyardSize(ctx.Game, ctx.Controller())) + "}"
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     cost,
				Question: "Rakshasa's Disdain — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
