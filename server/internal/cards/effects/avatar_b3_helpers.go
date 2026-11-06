package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// avatar_b3_helpers.go — shared bodies for Avatar: The Last Airbender,
// batch 3 (#4459). The other Avatar batches land beside this one, so a
// helper here is new to the catalog: earthbend as a modal bullet had
// only Dai Li Indoctrination's inline body before.

// EarthbendTheModesTarget is "Earthbend N." as a modal bullet's body
// (Sandbenders' Storm): the bullet's own target land, re-checked at
// resolution (CR 608.2b), earthbent with n counters through the one
// keyword-action primitive (Earthbend, CR 701.66a). Pair it with
// EarthbendTargets() as the bullet's target clause.
func EarthbendTheModesTarget(n int) func(item *game.StackItem, ctx *Context, occ int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		return Earthbend{Target: t.ID, N: n}.Apply(ctx)
	}
}
