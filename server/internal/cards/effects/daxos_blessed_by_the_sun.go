package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daxos, Blessed by the Sun — Legendary Enchantment Creature —
// Demigod {W}{W}, 2/* (EDHREC rank 1761):
//
//	"Daxos's toughness is equal to your devotion to white. (Each {W}
//	 in the mana costs of permanents you control counts toward your
//	 devotion to white.)
//	 Whenever another creature you control enters or dies, you gain
//	 1 life."
//
// The white devotion deck's two-drop. The toughness is a Layer 7a
// CDA (Tarmogoyf's shape) set to devotionTo(controller, "W") on
// every recompute — Daxos's own {W}{W} counts, so he is at least a
// 2/2 while he is on the battlefield; the printed "*" is 0 in the
// card data and the CDA overrides it. The lifegain is one trigger
// watching two kinds (Sun Titan's shape): another creature entering
// under the controller's control, or one they controlled dying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d3c7c96-056f-408e-a834-fa45a430d3d4",
		Name:         "Daxos, Blessed by the Sun",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Toughness = devotionTo(g, source.Controller, "W")
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16AnotherCreatureYouControlEnteredOrDied(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Daxos, Blessed by the Sun — you gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
