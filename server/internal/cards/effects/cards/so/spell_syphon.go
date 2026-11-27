package cardsso

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spell Syphon — Instant {1}{U} (Shadowmoor, #5591):
//
//	"Counter target spell unless its controller pays {1} for each blue
//	 permanent you control."
//
// "You" is Spell Syphon's controller and the count is read as it resolves
// (Spell Rupture's shape): the blue permanents they control, tokens and
// lands included, by their current colours. With none the spell's controller
// pays nothing and it is not countered. Being a pay-unless it is legal
// against a spell that can't be countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f30e9371-2c12-496a-80e1-40c3408defed",
		Name:         "Spell Syphon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			x := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.HasColor(ctx.TextColor("U")) {
					x++
				}
			}
			if x <= 0 {
				return nil
			}
			cost := "{" + strconv.Itoa(x) + "}"
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     cost,
				Question: "Spell Syphon — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
