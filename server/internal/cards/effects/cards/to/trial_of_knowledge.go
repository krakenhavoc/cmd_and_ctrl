package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trial of Knowledge — {3}{U} Enchantment: "When this enchantment
// enters, draw three cards, then discard a card. When a Cartouche you
// control enters, return this enchantment to its owner's hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6155908a-0079-42c7-ae81-326a6faccb89",
		Name:         "Trial of Knowledge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Trial of Knowledge — draw three cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 3) }),
			effects.WhenACartoucheYouControlEntersReturnThis("Trial of Knowledge"),
		},
	})
}
