package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coal Stoker — Creature — Elemental {3}{R}, 3/3 (Dominaria Remastered,
// #5948):
//
//	"When this creature enters, if you cast it from your hand, add {R}{R}{R}."
//
// An intervening "if" (CR 603.4): the trigger is only put on the stack
// when the Stoker was cast from the hand, and the condition is checked
// again as it resolves, exactly as Scion of Vitu-Ghazi does it
// (b30CastFromHand reads the cast event's origin zone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f3f7fb0-da18-467b-bf57-2ded458fda4d",
		Name:         "Coal Stoker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return b06SelfETB(ev, source, lki, g) && b30CastFromHand(g, source.InstanceID)
			}, "Coal Stoker — add {R}{R}{R}",
				Do(Instruction(func(ctx *Context) error {
					if !b30CastFromHand(ctx.Game, ctx.Source()) {
						return nil
					}
					return AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
				}))),
		},
	})
}
