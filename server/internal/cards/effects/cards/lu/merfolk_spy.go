package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Spy — Creature — Merfolk Rogue {U}, 1/1 (Magic 2014, S82 old sets):
//
//	"Islandwalk
//	 Whenever this creature deals combat damage to a player, that player reveals
//	 a card at random from their hand."
//
// Islandwalk is the keyword. The trigger is Friendly Fire's random reveal: the
// card is picked from the game's keyed random stream (so an undo rewinds it) and
// revealed to the table, and an empty hand reveals nothing. The player is the
// one the damage event named.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d004b481-a1d5-459d-8282-76baa9d13652",
		Name:            "Merfolk Spy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Merfolk Spy — that player reveals a card at random from their hand",
				effects.S82OldB400RevealRandomFromDamagedHand("Merfolk Spy")),
		},
	})
}
