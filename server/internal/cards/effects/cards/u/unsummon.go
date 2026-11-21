package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unsummon — "Return target creature to its owner's hand." The
// BounceToHand primitive looks the card's owner up from its
// current zone and routes there, so this is a one-liner.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "837182db-1bf3-4a2c-bd01-1af9d9873561",
		Name:         "Unsummon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BounceToHand{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
