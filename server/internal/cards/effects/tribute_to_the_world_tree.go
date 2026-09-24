package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tribute to the World Tree — Enchantment {G}{G}{G} (EDHREC rank
// 608):
//
//	"Whenever a creature you control enters, draw a card if its power
//	 is 3 or greater. Otherwise, put two +1/+1 counters on it."
//
// Green's Guardian Project that never whiffs: big creatures draw,
// small ones grow into big ones. One ETB trigger per creature; the
// entering creature is the trigger's event object and its power is
// read at RESOLUTION (CR 608.2h reads the current value), so a Giant
// Growth in response flips a counter into a draw, as in paper.
//
// A creature that has left the battlefield by the time the trigger
// resolves is read by its last-known power (CR 608.2h, #1379), so a big
// creature killed in response still draws. The counters half does
// nothing for it: last-known information is read, never written to, and
// a creature that left and came back is a new object the trigger never
// named.
func init() {
	Register(Spec{
		OracleID:     "72deedab-7c17-4505-aeca-4bc8596d80a5",
		Name:         "Tribute to the World Tree",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Tribute to the World Tree — draw, or two +1/+1 counters",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						entered, ok := ctx.TriggeringPermanent()
						if !ok {
							return nil
						}
						if entered.Power >= 3 {
							return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
						}
						if entered.Left {
							return nil
						}
						return AddCounter{Target: ctx.Trigger().Object.ID, Kind: "+1/+1", N: 2}.Apply(ctx)
					})
			},
		}},
	})
}
