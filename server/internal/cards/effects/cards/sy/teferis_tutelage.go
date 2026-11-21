package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teferi's Tutelage — Enchantment {2}{U} (Jumpstart):
//
//	"When this enchantment enters, draw a card, then discard a card.
//	 Whenever you draw a card, target opponent mills two cards."
//
// Sphinx's Tutelage without the repeat. The draw trigger fires once per
// card drawn (EventDrawCard carries one card, the drawer in Actor), each
// with its own "target opponent" chosen as it goes on the stack, and the
// enters loot's draw is a draw like any other, so it triggers the mill
// too. A library holding fewer than two mills what it has, and milling
// out never loses a player the game (CR 701.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6b94e03-61c0-4077-b3da-a70795e97278",
		Name:         "Teferi's Tutelage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Teferi's Tutelage — draw a card, then discard a card", effects.B35LootOne),
			effects.Targeting(effects.WheneverYouDraw("Teferi's Tutelage — target opponent mills two cards", effects.TargetPlayerMills(2)),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
