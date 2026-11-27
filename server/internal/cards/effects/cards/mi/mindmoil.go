package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindmoil — Enchantment {4}{R} (Ravnica: City of Guilds):
//
//	"Whenever you cast a spell, put the cards in your hand on the bottom of your
//	 library in any order, then draw that many cards."
//
// Arjun, the Shifting Flame's trigger on an enchantment. It goes on the stack
// above the spell being cast (the spell has already left your hand, so it is not
// among the cards). As it resolves your whole hand is the pile: you choose the
// order it goes on the bottom of your library in, and once it has been placed you
// draw as many cards as were put there. An empty hand does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41e4ba67-d038-4d0a-b740-709cb2350192",
		Name:         "Mindmoil",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, "Mindmoil — put your hand on the bottom of your library in any order, then draw that many cards",
				effects.OldB579HandToBottomThenDraw("Mindmoil")),
		},
	})
}
