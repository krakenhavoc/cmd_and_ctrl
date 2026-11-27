package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Clock — Artifact {4} (Saviors of Kamigawa, #5601):
//
//	"At the beginning of each player's upkeep, that player returns a
//	 permanent they control to its owner's hand unless they pay 2 life."
//
// The player whose upkeep it is, not the clock's controller, answers: pay 2
// life, or choose a permanent they control (the clock included) and return it.
// A player with no permanent is not asked; one with under 2 life cannot pay
// (CR 119.4) and returns one. The permanent goes back to its owner's hand, so
// a permanent they control but do not own returns to the other player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef987ddb-25f6-4895-8f83-045254745ea4",
		Name:         "Blood Clock",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Blood Clock — that player returns a permanent they control to its owner's hand unless they pay 2 life", effects.OldB586BloodClock),
		},
	})
}
