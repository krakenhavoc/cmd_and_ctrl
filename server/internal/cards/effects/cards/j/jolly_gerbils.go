package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jolly Gerbils — Creature — Hamster Citizen {1}{W}, 2/1 (Bloomburrow,
// #4467):
//
//	"Whenever you give a gift, draw a card."
//
// Watches game.EventGiftGiven, which the gift keyword emits when a gift's
// effect happens for the chosen opponent (CR 702.174d–i): a spell's as its
// resolution begins, a permanent's from its entry trigger. A gift that was
// only promised and never delivered (the spell was countered) is not given.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e22bf51-1b1b-459b-af78-91188b6e3f09",
		Name:         "Jolly Gerbils",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventGiftGiven, effects.ByYou, "Jolly Gerbils — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
