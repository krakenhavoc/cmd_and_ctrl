package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teeming Dragonstorm — Enchantment {3}{W} (Tarkir: Dragonstorm, #4463):
//
//	"When this enchantment enters, create two 2/2 white Soldier creature
//	 tokens.
//	 When a Dragon you control enters, return this enchantment to its
//	 owner's hand."
//
// The enters trigger makes two plain 2/2 white Soldiers. The Dragon
// trigger is the cycle's shared one (dragonstorm.go): any Dragon entering
// under your control, a token or a changeling included, returns the
// enchantment if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2154cc68-41a4-4e88-9fe0-e0150dc0be45",
		Name:         "Teeming Dragonstorm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Teeming Dragonstorm — create two 2/2 white Soldier creature tokens",
				effects.Do(effects.CreateToken{Template: teemingDragonstormSoldier(), N: 2})),
			effects.WhenADragonYouControlEntersReturnThis("Teeming Dragonstorm"),
		},
	})
}

// teemingDragonstormSoldier is the 2/2 white Soldier, a plain token the
// table has no row for.
func teemingDragonstormSoldier() game.Card {
	return game.Card{
		Name:      "Soldier",
		TypeLine:  "Token Creature — Soldier",
		Colors:    []string{"W"},
		Power:     2,
		Toughness: 2,
	}
}
