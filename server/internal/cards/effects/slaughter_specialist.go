package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ab431f93-73bb-49d9-9321-a0712f096ed7",
		Name:         "Slaughter Specialist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Slaughter Specialist — each opponent creates a 1/1 white Human token",
				Do(Instruction(func(ctx *Context) error {
					for _, p := range ctx.Game.Seats {
						if p == nil || p.Eliminated || p.ID == ctx.Controller() {
							continue
						}
						if err := (CreateToken{Controller: p.ID, Template: TokenCard("1/1 white Human"), N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}))),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) != source.Controller
			}, "Slaughter Specialist — put a +1/+1 counter on Slaughter Specialist",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
