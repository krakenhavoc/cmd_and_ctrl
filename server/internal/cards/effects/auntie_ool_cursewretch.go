package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Auntie Ool, Cursewretch — Legendary Creature — Goblin Warlock
// {1}{B}{R}{G}, 4/4 (Lorwyn Eclipsed Commander, #5368):
//
//	"Ward—Blight 2. (To blight 2, a player puts two -1/-1 counters on a
//	 creature they control.)
//	 Whenever one or more -1/-1 counters are put on a creature, draw a card
//	 if you control that creature. If you don't control it, its controller
//	 loses 1 life."
//
// Ward—Blight 2 is the ward trigger with a blight payment (WardBlight): the
// player who targeted may put two -1/-1 counters on a creature they control
// or the spell or ability is countered, and a player with no creature to
// blight can't pay (CR 701.68b). The second is one trigger per placement
// event, any creature and any placer (cmdr2B25MinusCountersPlacedOnACreature).
// "That creature's" controller is read as the ability resolves while it is
// still on the battlefield, and as the event left it otherwise.
//
// No simplification.
func init() {
	const label = "Auntie Ool — draw a card if you control that creature, otherwise its controller loses 1 life"
	Register(Spec{
		OracleID:     "e3cd53c1-86bd-4177-adb0-a1fa693b44c6",
		Name:         "Auntie Ool, Cursewretch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardBlight(2), "Auntie Ool — ward—blight 2"),
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return cmdr2B25MinusCountersPlacedOnACreature(ev, g)
				},
				Key: label,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, label)
					item.Params.Object = game.ObjectRef{ID: ev.Target}
					if c, ok := g.LookupCardForEffect(ev.Target); ok {
						item.Params.Player = c.Controller
					}
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					controller := item.Params.Player
					if c, ok := g.LookupCardForEffect(item.Params.Object.ID); ok && onBattlefield(g, c.InstanceID) {
						controller = c.Controller
					}
					if controller == item.Controller {
						return DrawCards{N: 1}.Apply(ctx)
					}
					if g.PlayerByIDForEffect(controller) == nil {
						return nil
					}
					return g.ChangePlayerLifeForEffect(ctx.Source(), controller, -1)
				},
			},
		},
	})
}
