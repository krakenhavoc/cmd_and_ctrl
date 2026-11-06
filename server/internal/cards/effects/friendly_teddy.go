package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Friendly Teddy — Artifact Creature — Bear Toy {2}, 2/2 (Duskmourn:
// House of Horror, #4466):
//
//	"When this creature dies, each player draws a card."
//
// A death trigger whose body is Scrawling Crawler's b05EachPlayerDraws:
// every player still in the game draws one, in turn order from the
// active player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02a8b849-e708-427c-92fe-9a41bb0ec7fe",
		Name:         "Friendly Teddy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Friendly Teddy — each player draws a card", func(g *game.Game, item *game.StackItem) error {
				return b05EachPlayerDraws(g, item, 1)
			}),
		},
	})
}
