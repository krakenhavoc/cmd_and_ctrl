package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirri's Guile — Enchantment {G}:
//
//	"At the beginning of your upkeep, you may look at the top three cards of
//	 your library, then put them back in any order."
//
// An optional upkeep trigger over the look-at-top primitive (Aven
// Fateshaper's): looking is not revealing, so only the controller sees the
// cards, and the order they put them back in is the new top of the library.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f89c0ee-b914-406c-8a3f-98424a52ae14",
		Name:         "Mirri's Guile",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Mirri's Guile — look at the top three cards of your library, then put them back in any order",
				effects.Do(effects.LookAtTop{N: 3})),
				"Look at the top three cards of your library and put them back in any order?"),
		},
	})
}
