package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Refute Destiny — Sorcery {1}{W} (Reality Fracture, tracker #2795):
//
//	"Exile target creature or planeswalker that's green or blue. Surveil
//	 1. (Look at the top card of your library. You may put it into your
//	 graveyard.)"
//
// Surveil only queues a prompt, but nothing follows it, so it is the last
// statement. Colour is read from the effective colours at announce and
// again at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a18b3f0-8a78-46d8-b60f-3cd634062f1b",
		Name:         "Refute Destiny",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target creature or planeswalker that's green or blue",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.Or(effects.OfColor("G"), effects.OfColor("U")))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard && ctx.IsTargetLegal(item.Targets[0]) {
				if err := (effects.ExileTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
