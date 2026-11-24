package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindslicer — Creature — Horror {2}{B}{B}, 4/3 (Dominaria Remastered):
//
//	"When this creature dies, each player discards their hand."
//
// Every player still in the game, the Mindslicer's controller included,
// discards their whole hand (DiscardWholeHand: no choice, an empty hand
// discards nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76c229d2-4bdd-4520-97dd-afb40f8c65ec",
		Name:         "Mindslicer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Mindslicer — each player discards their hand", effects.EachPlayerDiscardsTheirHand),
		},
	})
}
