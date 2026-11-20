package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Runed Servitor — Artifact Creature — Construct {2}, 2/2 (Rise of the
// Eldrazi, #5581):
//
//	"When this creature dies, each player draws a card."
//
// Friendly Teddy's dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e5e81e1-ab47-4f08-9c42-dba0adc5cf9a",
		Name:         "Runed Servitor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Runed Servitor — each player draws a card", func(g *game.Game, item *game.StackItem) error {
				return b05EachPlayerDraws(g, item, 1)
			}),
		},
	})
}
