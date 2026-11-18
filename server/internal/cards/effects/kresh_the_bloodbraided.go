package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kresh the Bloodbraided — Legendary Creature — Human Warrior
// {2}{B}{R}{G}, 3/3:
//
//	"Whenever another creature dies, you may put X +1/+1 counters on
//	 Kresh, where X is that creature's power."
//
// Elenda's death trigger (any other creature, anyone's, tokens
// included) with a "may". X is the dead creature's LAST-KNOWN power
// (CR 603.10, 608.2h): the layer-applied power it had on the
// battlefield, counters and anthems included, fixed when the trigger
// is built (b13LastKnownPower). Power of zero or less puts no counters
// on. The counters are put on Kresh as the ability resolves; a Kresh
// that is gone by then has nothing to put them on.
//
// No simplification.
func init() {
	const label = "Kresh the Bloodbraided — put X +1/+1 counters on it, X that creature's power"
	Register(Spec{
		OracleID:     "6343cf6f-b1e3-49ce-8933-8fe6c9dc83f3",
		Name:         "Kresh the Bloodbraided",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return AnotherCreatureDied(ev, source, ch, g)
				},
				Key: label,
				Build: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, label)
					item.Params.Amount = b13LastKnownPower(g, ev.CardID, lki)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Params.Amount <= 0 {
						return nil
					}
					ctx := NewContext(g, item)
					return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(ctx)
				},
			}, "Kresh the Bloodbraided — put +1/+1 counters on it?"),
		},
	})
}
