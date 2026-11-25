package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mount Velus Manticore — Enchantment Creature — Manticore {2}{R}{R}, 3/4
// (Modern Horizons 2):
//
//	"At the beginning of combat on your turn, you may discard a card. When you
//	 do, this creature deals X damage to any target, where X is the number of
//	 card types the discarded card has."
//
// "You may discard" is an up-to-one discard prompt; the reflexive trigger
// (CR 603.12) is made only once a card really was discarded, with X fixed as the
// number of card types (CR 205.2a) the discarded card has, and chooses its target
// as it goes on the stack. An artifact creature is 2, a land 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f434590f-3a5d-47dd-8ed0-d2a43a521f86",
		Name:         "Mount Velus Manticore",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Mount Velus Manticore — you may discard a card; when you do, damage equal to its card types to any target",
				effects.OldB036ManticoreAsk),
		},
	})
}
