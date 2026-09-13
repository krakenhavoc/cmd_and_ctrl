package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harvester of Souls — Creature — Demon {4}{B}{B}, 5/5 (EDHREC rank
// 1886):
//
//	"Deathtouch (Any amount of damage this deals to a creature is
//	 enough to destroy it.)
//	 Whenever another nontoken creature dies, you may draw a card."
//
// The aristocrats' card-draw Demon that reads every death at the
// table, not just yours. "Another" excludes its own death by ID;
// "nontoken" reads the dead card's printed type line, which is how
// every token in the catalog is stamped. The draw is optional, so
// the trigger prompts before it goes on the stack (CR 603.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5987ce77-10ad-4871-900a-5a005fcf4955",
		Name:            "Harvester of Souls",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17AnotherNontokenCreatureDied(ev, source, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Harvester of Souls — draw a card?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Harvester of Souls — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
