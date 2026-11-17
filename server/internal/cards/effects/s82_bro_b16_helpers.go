package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_bro_b16_helpers.go — shared pieces for The Brothers' War batch 16 (#5482).

// putCountersOnTheModesTargetStep is "put N counters of `kind` on target
// <thing>" as one step of a modal bullet (ModeSteps): it reads the bullet's
// own target and does nothing once that target is gone (CR 608.2b). Konstrari
// Charm's and Gix's Command's first bullets share it, each followed by their
// own "it gains <keyword> until end of turn" step, which waits for a prompt
// the counters ask (a CR 616 order; CR 608.2c, #4155).
func putCountersOnTheModesTargetStep(kind string, n int) func(ctx *Context) error {
	return func(ctx *Context) error {
		t, ok := ModeTarget(ctx, ctx.Occurrence())
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		return AddCounter{Target: t.ID, Kind: kind, N: n}.Apply(ctx)
	}
}
