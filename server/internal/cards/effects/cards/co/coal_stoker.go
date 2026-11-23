package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "7f3f7fb0-da18-467b-bf57-2ded458fda4d",
		Name:         "Coal Stoker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.B06SelfETB(ev, source, lki, g) && effects.B30CastFromHand(g, source.InstanceID)
			}, "Coal Stoker — add {R}{R}{R}",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if !effects.B30CastFromHand(ctx.Game, ctx.Source()) {
						return nil
					}
					return effects.AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
				}))),
		},
	})
}
