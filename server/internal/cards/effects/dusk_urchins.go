package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dusk Urchins — Creature — Ouphe {2}{B}, 4/3 (Lorwyn Eclipsed Commander,
// #5368):
//
//	"Whenever this creature attacks or blocks, put a -1/-1 counter on it.
//	 When this creature dies, draw a card for each -1/-1 counter on it."
//
// Two triggers. The first has no target and puts one counter on the Urchins
// itself (a block is announced once, however many creatures it blocks:
// selfBlocksOnce). The second reads the counters the Urchins last had on the
// battlefield (CR 603.10a, LastKnownCountersForEffect), because the card in
// the graveyard has none, so one that is killed by its own attack trigger
// still draws for the counter that shrank it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c2cc166-dc59-44ce-86d5-ed6013b0bf67",
		Name:         "Dusk Urchins",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
				},
				"Dusk Urchins — put a -1/-1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterMinusOne, N: 1}.Apply(NewContext(g, item))
				}),
			{
				Watches:   []game.EventKind{game.EventLTB},
				AppliesTo: ThisDied,
				Key:       "Dusk Urchins — draw a card for each -1/-1 counter on it",
				// The counters it last had are read at the event (CR 603.10a),
				// before the resolving item could look for them: the log's
				// snapshot is gone by then. They ride the item as its amount.
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Dusk Urchins — draw a card for each -1/-1 counter on it")
					item.Params.Amount = g.LastKnownCountersForEffect(ev.CardID)[game.CounterMinusOne]
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{N: item.Params.Amount}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
