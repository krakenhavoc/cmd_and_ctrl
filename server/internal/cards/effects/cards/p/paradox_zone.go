package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paradox Zone — Enchantment {4}{G} (Commander 2021, #5504):
//
//	"This enchantment enters with a growth counter on it.
//	 At the beginning of your end step, double the number of growth
//	 counters on this enchantment. Then create a 0/0 green and blue
//	 Fractal creature token. Put X +1/+1 counters on it, where X is the
//	 number of growth counters on this enchantment."
//
// The counter is an entry replacement (CR 614.1c). Doubling adds as many
// growth counters as there are (CR 122.1 doubling reads the count as the
// ability resolves, so counter-doublers apply to the added ones), then the
// Fractal is the create-then-put shape (SocB8FractalWithCountersSteps):
// the token is created first and X counters are put on it, X being the
// growth counters AFTER the doubling. An enchantment that has left makes
// the Fractal with X read last-known (0), which dies as a 0/0.
//
// No simplification.
func init() {
	const key = "paradox-zone/fractal"
	steps := append([]effects.Applier{effects.Instruction(func(ctx *effects.Context) error {
		n := growthCounters(ctx)
		if n <= 0 || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
			return nil
		}
		return effects.AddCounter{Target: ctx.Source(), Kind: "growth", N: n}.Apply(ctx)
	})}, effects.SocB8FractalWithCountersSteps(key, growthCounters)...)
	effects.Register(effects.Spec{
		OracleID:     "2fff8441-1f5a-423b-ab98-93016266f5c5",
		Name:         "Paradox Zone",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters("growth", 1, "Paradox Zone enters with a growth counter on it")},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Paradox Zone — double the growth counters, then create a Fractal with that many +1/+1 counters", effects.Do(steps...)),
		},
	})
}

// growthCounters is the growth counters on this enchantment as it is now.
func growthCounters(ctx *effects.Context) int {
	c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok {
		return 0
	}
	return c.Counters["growth"]
}
