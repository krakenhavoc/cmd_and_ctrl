package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honden of Life's Web — Legendary Enchantment — Shrine {4}{G}
// (Champions of Kamigawa, #5603):
//
//	"At the beginning of your upkeep, create a 1/1 colorless Spirit
//	 creature token for each Shrine you control."
//
// The Shrines are counted as the trigger resolves (CR 608.2h); the
// tokens are the shared 1/1 colorless Spirit row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a1fdfcc6-4b3c-40f8-a9e4-dff9ace74a59",
		Name:         "Honden of Life's Web",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chkB1Honden("Honden of Life's Web — create a 1/1 colorless Spirit creature token for each Shrine you control",
				func(g *game.Game, item *game.StackItem, shrines int) error {
					return CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: shrines}.Apply(NewContext(g, item))
				}),
		},
	})
}
