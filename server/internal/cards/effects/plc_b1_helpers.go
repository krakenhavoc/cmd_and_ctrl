package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// plc_b1_helpers.go — small bodies shared by the Planar Chaos cards of
// batch 1 (#5595).

// plcGrantTheModesTarget is "target creature gains <keyword> until end of
// turn" as a modal bullet's body (Midnight, Piracy and Evolution Charm):
// this occurrence's own target, while it is still legal (CR 608.2b).
func PlcGrantTheModesTarget(label, keyword string) func(*game.StackItem, *Context, int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{keyword}, Label: label}.Apply(ctx)
	}
}
