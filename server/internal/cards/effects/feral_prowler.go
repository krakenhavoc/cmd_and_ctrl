package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feral Prowler — Creature — Cat {1}{G}, 1/3 (Hour of Devastation, #5534):
//
//	"When this creature dies, draw a card."
//
// A dies trigger (CR 700.4), Pelakka Wurm's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bebba2d-867f-4ed1-a308-904fac32c8b6",
		Name:         "Feral Prowler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Feral Prowler — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
