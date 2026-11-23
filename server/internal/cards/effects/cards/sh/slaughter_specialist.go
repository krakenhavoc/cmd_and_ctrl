package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slaughter Specialist — Creature — Vampire Warrior {1}{B}, 3/3 (Innistrad:
// Midnight Hunt, #5497):
//
//	"When this creature enters, each opponent creates a 1/1 white Human
//	 creature token.
//	 Whenever a creature an opponent controls dies, put a +1/+1 counter on
//	 this creature."
//
// Each opponent is the token's controller (CreateToken with that player).
// The counter is Yahenni's: a creature that left the battlefield under an
// opponent's control, read as it last existed (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab431f93-73bb-49d9-9321-a0712f096ed7",
		Name:         "Slaughter Specialist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.Self, "Slaughter Specialist — each opponent creates a 1/1 white Human token",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					for _, p := range ctx.Game.Seats {
						if p == nil || p.Eliminated || p.ID == ctx.Controller() {
							continue
						}
						if err := (effects.CreateToken{Controller: p.ID, Template: effects.TokenCard("1/1 white Human"), N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}))),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) != source.Controller
			}, "Slaughter Specialist — put a +1/+1 counter on Slaughter Specialist",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
