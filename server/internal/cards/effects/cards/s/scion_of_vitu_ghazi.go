package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scion of Vitu-Ghazi — Creature — Elemental {3}{W}{W}, 4/4:
//
//	"When this creature enters, if you cast it from your hand, create
//	 a 1/1 white Bird creature token with flying, then populate.
//	 (Create a token that's a copy of a creature token you control.)"
//
// An intervening-if (CR 603.4): "you cast it from your hand" is read
// off the event log when the trigger fires and again as it resolves
// (b30CastFromHand), exactly as Wakening Sun's Avatar does. A
// reanimated, blinked or commander-zone Scion makes nothing. The Bird
// is made first, so it is a candidate for the populate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a34c4d5c-6ecb-4da2-8734-a2f3d35cfe8b",
		Name:         "Scion of Vitu-Ghazi",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.B06SelfETB(ev, source, lki, g) && effects.B30CastFromHand(g, source.InstanceID)
			}, "Scion of Vitu-Ghazi — a Bird, then populate",
				// Populate waits for a prompt the token asks (a CR 616 order; CR 608.2c, #4155).
				// Whether it was cast from hand does not change while the trigger resolves.
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						if !effects.B30CastFromHand(ctx.Game, ctx.Source()) {
							return nil
						}
						return effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: 1}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						if !effects.B30CastFromHand(ctx.Game, ctx.Source()) {
							return nil
						}
						return effects.Populate{}.Apply(ctx)
					}),
				)),
		},
	})
}
