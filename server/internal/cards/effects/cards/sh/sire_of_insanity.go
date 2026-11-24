package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sire of Insanity — Creature — Demon {4}{B}{R}, 6/4 (Secret Lair Drop):
//
//	"At the beginning of each end step, each player discards their hand."
//
// The trigger is on every player's end step; each player discards the whole
// hand (no choice), Mindslicer's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b55d8cd6-037d-4d7c-90ec-51249d34b02e",
		Name:         "Sire of Insanity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Sire of Insanity — each player discards their hand", effects.EachPlayerDiscardsTheirHand),
		},
	})
}
