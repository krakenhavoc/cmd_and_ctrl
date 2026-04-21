package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unsummon — "Return target creature to its owner's hand." The
// BounceToHand primitive looks the card's owner up from its
// current zone and routes there, so this is a one-liner.
func init() {
	Register(Spec{
		OracleID:   "837182db-1bf3-4a2c-bd01-1af9d9873561",
		Name:       "Unsummon",
		TargetMode: "creature",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BounceToHand{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
