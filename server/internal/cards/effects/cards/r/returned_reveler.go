package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Returned Reveler — Creature — Zombie Satyr {1}{B}, 1/3 (Journey into Nyx, #5558):
//
//	"When this creature dies, each player mills three cards."
//
// Koma's Faithful's dies trigger: every player mills three, in APNAP
// order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f90df69a-d53f-44b1-833a-5223d2e6c296",
		Name:         "Returned Reveler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Returned Reveler — each player mills three cards", effects.EachPlayerMillsThree),
		},
	})
}
