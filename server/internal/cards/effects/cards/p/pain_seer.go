package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pain Seer — {1}{B} Creature — Human Wizard 2/2 (Born of the Gods):
//
//	"Inspired — Whenever this creature becomes untapped, reveal the top card of your
//	 library and put that card into your hand. You lose life equal to that card's mana
//	 value."
//
// Dark Confidant's flip: the reveal is not a draw, the mana value is read before the
// move, a land costs nothing and an empty library does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "910036e7-ae27-4adb-a314-d9760bce1eb0",
		Name:         "Pain Seer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self, "Pain Seer — reveal the top card of your library and put it into your hand; lose life equal to its mana value (inspired)", effects.DarkConfidantFlip),
		},
	})
}
