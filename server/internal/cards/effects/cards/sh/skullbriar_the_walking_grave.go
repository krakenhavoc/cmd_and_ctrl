package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const skullbriarOracleID = "088ac883-6f44-4046-bf19-81c7fd952e55"

// Skullbriar, the Walking Grave — Legendary Creature — Zombie Elemental {B}{G},
// 1/1 (Double Masters 2022):
//
//	"Haste
//	 Whenever Skullbriar deals combat damage to a player, put a +1/+1 counter
//	 on it.
//	 Counters remain on Skullbriar as it moves to any zone other than a
//	 player's hand or library."
//
// The counters rule is the engine's (game.RegisterCountersRemain, read by
// MoveCard): the +1/+1 counters, and any others, stay on the card through the
// graveyard, exile, the stack and back onto the battlefield, and are lost the
// moment it goes to a hand or a library.
//
// No simplification.
func init() {
	game.RegisterCountersRemain(skullbriarOracleID)
	effects.Register(effects.Spec{
		OracleID:        skullbriarOracleID,
		Name:            "Skullbriar, the Walking Grave",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Skullbriar, the Walking Grave — put a +1/+1 counter on it", effects.PlusOneCountersOnThis(1)),
		},
	})
}
