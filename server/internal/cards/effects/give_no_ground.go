package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Give No Ground — Instant {3}{W}:
//
//	"Target creature gets +2/+6 until end of turn and can block any
//	 number of creatures this turn."
//
// One sentence, one effect (CR 613.7): the +2/+6 (layer 7c) and "any
// number" (layer 6, #1715) are two mods of ONE data record pinned to
// the target, sharing a timestamp and ending together at cleanup.
func init() {
	Register(Spec{
		OracleID:     "bae18a46-00e0-4b3e-8fdb-44cc61f30f8e",
		Name:         "Give No Ground",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			mods := append([]game.Mod{game.ModifyPTMod(2, 6)}, blockCapacityMods(0, true)...)
			return untilEndOfTurn(ctx, ts[0].ID, nil, "Give No Ground — +2/+6, can block any number of creatures", mods...)
		},
	})
}
