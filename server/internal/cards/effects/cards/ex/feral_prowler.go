package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Prowler — Creature — Cat {1}{G}, 1/3 (Hour of Devastation, #5534):
//
//	"When this creature dies, draw a card."
//
// A dies trigger (CR 700.4), Pelakka Wurm's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bebba2d-867f-4ed1-a308-904fac32c8b6",
		Name:         "Feral Prowler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Feral Prowler — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
