package cardsth

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thassa's Rebuff — Instant {1}{U} (Born of the Gods, S82 OLD b379, #5559):
//
//	"Counter target spell unless its controller pays {X}, where X is your
//	 devotion to blue."
//
// Mindstatic's counter-unless-paid with the price counted as the spell
// resolves (CR 700.5); with no devotion the price is {0}, which is always
// paid, so nothing is countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc5d9899-1a1c-4779-b9f0-b00f2c233699",
		Name:         "Thassa's Rebuff",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			x := effects.DevotionTo(ctx.Game, ctx.Controller(), ctx.TextColor("U"))
			if x <= 0 {
				return nil
			}
			cost := "{" + strconv.Itoa(x) + "}"
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     cost,
				Question: "Thassa's Rebuff — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
