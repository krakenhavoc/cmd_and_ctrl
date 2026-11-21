package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// the trigger prompts before it goes on the stack (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5987ce77-10ad-4871-900a-5a005fcf4955",
		Name:            "Harvester of Souls",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B17AnotherNontokenCreatureDied(ev, source, g)
			}, "Harvester of Souls — draw a card", effects.Do(effects.DrawCards{N: 1})), "Harvester of Souls — draw a card?", 1),
		},
	})
}
