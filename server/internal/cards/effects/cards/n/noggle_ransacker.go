package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noggle Ransacker — Creature — Noggle Rogue {2}{U/R}, 2/1 (Planechase 2012,
// S82 OLD, #5569):
//
//	"When this creature enters, each player draws two cards, then discards a
//	 card at random."
//
// Each player in turn order from the controller draws two and then discards one
// card at random from the hand that leaves (CR 701.9b); a player with an empty
// hand after the draw discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "162aed84-f87f-4bd4-9cac-5008766bdfb7",
		Name:         "Noggle Ransacker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Noggle Ransacker — each player draws two cards, then discards a card at random", effects.S82OldB431NoggleRansackerBody),
		},
	})
}
