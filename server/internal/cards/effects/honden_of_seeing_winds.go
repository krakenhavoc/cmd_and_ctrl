package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honden of Seeing Winds — Legendary Enchantment — Shrine {4}{U}
// (Champions of Kamigawa, #5603):
//
//	"At the beginning of your upkeep, draw a card for each Shrine you
//	 control."
//
// The Shrines are counted as the trigger resolves (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45beca46-480c-4771-ba46-78b6ac7d5c2c",
		Name:         "Honden of Seeing Winds",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chkB1Honden("Honden of Seeing Winds — draw a card for each Shrine you control",
				func(g *game.Game, item *game.StackItem, shrines int) error {
					return DrawCards{N: shrines}.Apply(NewContext(g, item))
				}),
		},
	})
}
