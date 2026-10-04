package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valgavoth, Harrower of Souls — Legendary Creature — Elder Demon
// {2}{B}{R}, 4/4:
//
//	"Flying
//	 Ward—Pay 2 life.
//	 Whenever an opponent loses life for the first time during each of
//	 their turns, put a +1/+1 counter on Valgavoth and draw a card."
//
// Flying rides PrintedKeywords and the ward is the shared WardLife.
// The trigger watches the two shapes a life loss takes
// (b04OpponentLostLife: a negative EventChangeLife, or damage to a
// player, which writes the life total directly) and passes only when
//
//   - the player losing the life is the active player ("during each of
//     THEIR turns" — an opponent losing life on somebody else's turn
//     is not it), and
//   - this loss is the first of that turn: the per-turn tally already
//     holds exactly this event's amount (b18LifeLostThisTurn), so any
//     earlier loss makes it larger.
//
// Life lost to a prevented or zero-point event never emits an event,
// so it can't start the turn's count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5401439c-4068-4490-8d53-8ffd0026dd7d",
		Name:            "Valgavoth, Harrower of Souls",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(2), "Valgavoth, Harrower of Souls — ward, pay 2 life"),
			{
				Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					lost, ok := b04OpponentLostLife(ev, source.Controller, g)
					if !ok {
						return false
					}
					active := g.Seats[g.Turn.ActiveSeat]
					if active == nil || active.ID != ev.Target {
						return false
					}
					return b18LifeLostThisTurn(g, ev.Target) == lost
				},
				Key: "Valgavoth, Harrower of Souls — put a +1/+1 counter on it and draw a card",
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := b35PutCounterOnSelf(g, item); err != nil {
						return err
					}
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
