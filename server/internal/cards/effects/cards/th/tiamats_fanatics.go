package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tiamat's Fanatics — Creature — Dragon Warrior {4}{R}, 4/3 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Haste
//	 Myriad (Whenever this creature attacks, for each opponent other than defending
//	 player, you may create a token copy that's tapped and attacking that player or a
//	 planeswalker they control. Exile the tokens at end of combat.)"
//
// Haste is printed and myriad is the shared trigger (CR 702.116).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6eef7583-0716-49b5-b70e-8feb0e7c99fb",
		Name:            "Tiamat's Fanatics",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{effects.GrantedMyriad("Tiamat's Fanatics")},
	})
}
