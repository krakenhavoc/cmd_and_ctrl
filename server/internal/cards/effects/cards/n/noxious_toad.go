package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noxious Toad — Creature — Frog {2}{B}, 1/1 (Portal, #5630):
//
//	"When this creature dies, each opponent discards a card."
//
// Cackling Fiend's "each opponent discards a card", on a dies
// trigger, which looks back in time (CR 603.10a). Each opponent
// chooses the card they discard (CR 701.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fdc484e-b3c3-4f4e-99a1-26a1134aa1cd",
		Name:         "Noxious Toad",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Noxious Toad — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
	})
}
