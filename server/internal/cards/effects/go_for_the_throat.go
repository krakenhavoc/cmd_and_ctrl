package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Go for the Throat — Instant {1}{B} (EDHREC rank 375):
//
//	"Destroy target nonartifact creature."
//
// Doom Blade with the clause swapped from colour to type. Artifact()
// reads the post-layer type set, so a creature animated into an
// artifact is correctly not a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "2f092562-9e17-43cd-aeb8-d0567f99363e",
		Name:     "Go for the Throat",
		Targets:  TargetCreature("target nonartifact creature", Not(Artifact())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
