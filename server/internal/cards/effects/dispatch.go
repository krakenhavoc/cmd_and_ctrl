package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dispatch — Instant {W} (EDHREC rank 455):
//
//	"Tap target creature.
//	 Metalcraft — If you control three or more artifacts, exile that
//	 creature."
//
// One-mana exile in any deck with a Sol Ring, a Signet and a
// Treasure — which in Commander is most of them. Metalcraft is a
// resolution-time check, not a cast restriction: the spell always
// taps, and exiles only if the artifact count is three or more AS IT
// RESOLVES (a Treasure cracked in response can turn it off). The
// count reads post-layer types, so an animated artifact and a
// Mycosynth Lattice board both count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "133c99c0-3652-410f-8100-68015a47af9f",
		Name:     "Dispatch",
		Targets:  TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (TapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if b03ArtifactsControlled(ctx.Game, ctx.Controller()) < 3 {
				return nil
			}
			return ExileTarget{Target: target}.Apply(ctx)
		},
	})
}
