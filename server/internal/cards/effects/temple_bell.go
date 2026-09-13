package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temple Bell — Artifact {3} (EDHREC rank 1893):
//
//	"{T}: Each player draws a card."
//
// The group-hug rock, and the other half of the Mind Over Matter
// loop. One tap ability; every seated player draws, in APNAP order
// from the active player (b05EachPlayerDraws), so a table full of
// "whenever an opponent draws" triggers sees the draws in the
// printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc8032e9-c83a-4cf9-92e1-b1d2d9642695",
		Name:         "Temple Bell",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: Each player draws a card",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b05EachPlayerDraws(g, item, 1)
			},
		}},
	})
}
