package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Descent into Madness — Enchantment {3}{B}{B}:
//
//	"At the beginning of your upkeep, put a despair counter on this
//	 enchantment, then each player exiles X permanents they control and/or
//	 cards from their hand, where X is the number of despair counters on this
//	 enchantment."
//
// X is what the counter count LANDED at (a counter doubler can pause the
// placement on a CR 616 prompt), read after the placement. A Descent that
// has left the battlefield by resolution counts as it last existed.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88d42e0e-3921-4873-a53b-146220341f76",
		Name:         "Descent into Madness",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Key: "Descent into Madness — put a despair counter, then each player exiles X permanents and/or cards from hand",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Descent into Madness — put a despair counter, then each player exiles X permanents and/or cards from hand")
				item.Params.Amount = source.Counters["despair"]
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				lastKnown := item.Params.Amount
				if !effects.B15OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return effects.EachPlayerExilesPermanentsAndOrHandCards(effects.NewContext(g, item), lastKnown)
				}
				then := effects.NewContext(g, item).ItemTail(descentIntoMadnessExileTail)
				then.N = lastKnown
				return g.AddCounterByKeyedForEffect(uuid.Nil, item.SourceCardID, "despair", 1, then)
			},
		}},
	})
}

var descentIntoMadnessExileTail = effects.CardTail("descent-into-madness/exile", func(ctx *effects.Context, o game.TailOutcome) error {
	n := o.Tail.N
	if c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID); ok {
		n = c.Counters["despair"]
	}
	return effects.EachPlayerExilesPermanentsAndOrHandCards(ctx, n)
})
