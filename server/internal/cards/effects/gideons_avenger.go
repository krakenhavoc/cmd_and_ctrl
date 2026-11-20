package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gideon's Avenger — Creature — Human Soldier {1}{W}{W}, 2/2 (Magic 2012,
// #5573):
//
//	"Whenever a creature an opponent controls becomes tapped, put a
//	 +1/+1 counter on this creature."
//
// Rhoda, Geist Avenger's trigger: a creature an opponent controls becoming
// tapped, on the battlefield, puts a +1/+1 counter on it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ca79852-d90d-47b0-8427-4da90779e205",
		Name:         "Gideon's Avenger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventTapCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller != source.Controller && g.Battlefield.Contains(ev.CardID)
			}, "Gideon's Avenger — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
