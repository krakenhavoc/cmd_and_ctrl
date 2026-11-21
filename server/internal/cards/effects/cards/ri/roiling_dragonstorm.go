package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roiling Dragonstorm — Enchantment {1}{U} (Tarkir: Dragonstorm, #4463):
//
//	"When this enchantment enters, draw two cards, then discard a card.
//	 When a Dragon you control enters, return this enchantment to its
//	 owner's hand."
//
// The enters trigger is Mask of Memory's loot body
// (maskOfMemoryDrawThenDiscard): both draws land before the discard is
// chosen, so a drawn card may be the discard. The Dragon
// trigger is the cycle's shared one (dragonstorm.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6badc583-fabb-4d63-b70c-e7380acafc23",
		Name:         "Roiling Dragonstorm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Discards: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Roiling Dragonstorm — draw two cards, then discard a card",
				effects.Do(effects.Instruction(effects.MaskOfMemoryDrawThenDiscard))),
			effects.WhenADragonYouControlEntersReturnThis("Roiling Dragonstorm"),
		},
	})
}
