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
// entering creature rides the trigger as a copied ID and its power is
// read at RESOLUTION (CR 608.2h reads the current value), so a Giant
// Growth in response flips a counter into a draw, as in paper.
//
// Sandbox simplification: a creature that has left the battlefield
// by the time the trigger resolves gets nothing — no draw, no
// counters. Printed, the draw half would still use last-known power;
// that needs the LKI the dies-path keeps and the ETB path does not.
// Weaker than printed, and a corner.
func init() {
	Register(Spec{
		OracleID:     "72deedab-7c17-4505-aeca-4bc8596d80a5",
		Name:         "Tribute to the World Tree",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A creature that has already left the battlefield when the trigger resolves gets nothing, neither a card nor the counters."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				entered := ev.CardID
				return game.NewTriggeredItem(source, "Tribute to the World Tree — draw, or two +1/+1 counters",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						c, ok := g.LookupCardForEffect(entered)
						if !ok || !g.Battlefield.Contains(entered) {
							return nil
						}
						if c.CurrentPower() >= 3 {
							return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
						}
						return AddCounter{Target: entered, Kind: "+1/+1", N: 2}.Apply(ctx)
					})
			},
		}},
	})
}
